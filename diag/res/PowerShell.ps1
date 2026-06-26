param([string] $state = "unknown", [string] $outputPath = "C:\dptrace", [int] $verbosity = 4)

# INIT

$ErrorActionPreference = 'Continue'
$ConfirmPreference = 'None'
$ProgressPreference = 'SilentlyContinue'
$PSDefaultParameterValues['Invoke-WebRequest:UseBasicParsing'] = $true
$dpPath = Get-ItemPropertyValue HKLM:\SOFTWARE\DigitalPersona DigitalPersonaPath -ErrorAction SilentlyContinue
$wmcPath = Join-Path $dpPath "Web Management Components"
$start = $state -eq "start"
$stop = $state -eq "stop"

function Log {
  [CmdletBinding()]
  param (
    [Parameter(Position = 0)]
    $Data
  )
  "$([datetime]::UtcNow) $Data" | Out-File "$outputPath\info_trace.txt" -Append 
}

# TRACE SESSION

Log "----BEGIN----"
Log "[header] state $($state)"
Log "[header] outputPath $($outputPath)"
Log "[header] verbosity $($verbosity)"
Log "[header] date utc $([datetime]::UtcNow -f 'u')"

# SAML

$satosaProxyConfPath = "$dpPath\Web Management Components\DP STS\DPSaml\proxy_conf.yaml"
if (Test-Path $satosaProxyConfPath) {

  Log "[satosa] begin"

  if ($start) {
    (Get-Content $satosaProxyConfPath) -replace "ERROR", "DEBUG" | Set-Content $satosaProxyConfPath
    Log "[satosa] start replace"
  }

  if ($stop) {
    (Get-Content $satosaProxyConfPath) -replace "DEBUG", "ERROR" | Set-Content $satosaProxyConfPath
    Log "[satosa] stop replace"
    Copy-Item -Path "$env:ProgramData\DigitalPersona\SAML\satosa.log" -Destination $outputPath -ErrorAction SilentlyContinue
    Log "[satosa] stop copy"
    Remove-Item -Path "$env:ProgramData\DigitalPersona\SAML\satosa.log" -Force -ErrorAction SilentlyContinue
    Log "[satosa] stop remove"
  }
    
  Log "[satosa] end"
}

# WMC CONFIGS

if (Test-Path $wmcPath) {
  if ($stop) {
    Log "[config] begin"
    Get-ChildItem -Path $wmcPath -Include appsettings.json, metadata*.xml, *.config, *.yaml -Recurse `
    | ForEach-Object { 
      Copy-Item -Path $_.FullName -Destination "$outputPath\wmc$($_.FullName.Replace($wmcPath, '').Replace('\', '_'))" -Force
    }
    Log "[config] end"
  }
}

# DP REGISTRY

if ($stop) {
  Log "[reg] begin" 
  Log "[reg] apps"
  reg export "HKLM\SOFTWARE\DigitalPersona\Applications" "$outputPath\reg_applications.txt" /y | Out-Null      
  Log "[reg] policies"
  reg export "HKLM\SOFTWARE\DigitalPersona\Policies" "$outputPath\reg_policies_default.txt" /y | Out-Null 
  Log "[reg] dp"
  reg export "HKLM\SOFTWARE\Policies\DigitalPersona" "$outputPath\reg_policies_set.txt" /y | Out-Null
  Log "[reg] end"
}

# INSTALLED PRODUCTS

if ($stop) {  
  Log "[products] begin"
  Get-ChildItem "HKLM:\SOFTWARE\DigitalPersona\Products" -ErrorAction SilentlyContinue `
  | Get-ItemProperty `
  | Select-Object -Property "*Cap*", "Version" `
  | Format-List * `
  | Out-File "$outputPath\info_products.txt"
  Log "[products] end"
}

# WMC DNS

if ($stop -and ($verbosity -ge 7)) {
  Log "[dns] begin"
  $nslookupPath = "$outputPath\info_nslookup.txt"
  "" | Out-File $nslookupPath
  Get-ChildItem HKLM:\SOFTWARE\DigitalPersona\Applications\ConfigurationTool `
  | ForEach-Object { ([System.Uri](Get-ItemPropertyValue $_.PSPath RootUrl)).DnsSafeHost } `
  | ForEach-Object {
    Log "[dns] $($_)"
    $_ | Out-File $nslookupPath -Append   
    nslookup $_ | Out-File $nslookupPath -Append
  }
  Log "[dns] end"
}

# WINDOWS VERSION

if ($stop) {  
  Log "[winver] begin"
  cmd /c ver | Out-File "$outputPath\info_windows.txt"
  Log "[winver] end"
}

# DOTNET

if ($stop) {
  Log "[dotnet] begin"
  dotnet --info | Out-File "$outputPath\info_dotnet.txt"
  Log "[dotnet] end"
}

# DOMAIN TRUST

if ($stop) {
  Log "[trust] begin"
  & { Test-ComputerSecureChannel -Verbose } 4>&1 3>&1 2>&1 > "$outputPath\info_domaintrust.txt"
  Log "[trust] end"
}

# IP

if ($stop -and ($verbosity -ge 7)) {
  Log "[ip] begin"
  ipconfig | Out-File "$outputPath\info_ipconfig.txt"
  Log "[ip] end"
}

# SYSTEM INFO

if ($stop -and ($verbosity -ge 7)) {
  Log "[sysinfo] begin"
  systeminfo | Out-File "$outputPath\info_systeminfo.txt"
  Log "[sysinfo] end"
}

# SERVER FEATURES

if ($stop -and ($verbosity -ge 5)) {
  try {
    Log "[features] begin"
    Get-WindowsFeature | Sort-Object Name | Format-Table Name, Installed > "$outputPath\info_features.txt" 
    Log "[features] end" 
  }
  catch {      
    Log "[features] error $($_.Exception.Message)"  
  }
}

# EVENT LOGS

if ($stop -and ($verbosity -ge 5)) {  
  Log "[events] begin"
  wevtutil epl "Application" "$outputPath\appication.evtx" /ow:true
  Log "[events] end"
}

# IIS

if ($stop) {
  try {    
    Log "[iss] begin"
    Get-IISSite | ForEach-Object {
      Log "[iss] site $($_.Name)"
      [pscustomobject]@{
        Id              = $_.Id
        Name            = $_.Name
        ServerAutoStart = $_.ServerAutoStart
        State           = $_.State
        Bindings        = $_.Bindings | ForEach-Object {           
          Log "[iss] binding $($_.BindingInformation)"
          [pscustomobject]@{
            BindingInformation   = $_.BindingInformation
            CertificateHash      = ($_.CertificateHash | ForEach-Object { "{0:X2}" -f $_ }) -join ""
            CertificateStoreName = $_.CertificateStoreName
            EndPoint             = $_.EndPoint.ToString()
            Host                 = $_.Host
            Protocol             = $_.Protocol
          }
        }
        Applications    = $_.Applications | ForEach-Object {           
          Log "[iss] app $($_.Path)"
          [pscustomobject]@{
            ApplicationPoolName = $_.ApplicationPoolName
            EnabledProtocols    = $_.EnabledProtocols
            Path                = $_.Path
            VirtualDirectories  = $_.VirtualDirectories | ForEach-Object {
              [pscustomobject]@{
                PhysicalPath = $_.physicalPath                  
              }
            }    
          }
        }                
      }
    } | ConvertTo-Json -Depth 5 | Out-File "$outputPath\iis_sites.json"
    Log "[iss] end"
  }
  catch {            
    Log "[iss] error $($_.Exception.Message)"
  }
}

# IIS LOGS

if (Test-Path $wmcPath) {
  if ($stop) {    
    Log "[iss] log begin"
    Get-ChildItem -Path C:\inetpub\logs\LogFiles -Include *.log -Recurse `
    | Sort-Object { $_.LastWriteTime } `
    | Select-Object -Last 15 `
    | ForEach-Object { 
      Log "[iss] log copy $($_.FullName)"
      Copy-Item $_.FullName -Destination "$outputPath\iis_$($_.Directory.Name)_$($_.Name)" 
    }
    Log "[iss] log end"
  }
}

# WMC CERTS LIST

if (Test-Path $wmcPath) {
  if ($stop) {
    Log "[cert] info begin"
    Get-ChildItem Cert:\LocalMachine\My `
    | Format-List Thumbprint, Subject, NotAfter, DnsNameList `
    | Out-File "$outputPath\info_certs.txt"
    Log "[cert] info end"
  }
}

# CONFIGURATION TOOL

if ($stop) {
  Log "[conf log] begin"
  Get-ChildItem $env:TEMP -Include dpwmc_deployment*.log -Recurse `
  | Sort-Object { $_.LastWriteTime } `
  | Select-Object -Last 5 `
  | Copy-Item -Destination $outputPath
  Log "[conf log] end"
}

# ADFS

if ($stop -and ($verbosity -ge 5)) {
  try {
    Log "[adfs] begin"
    $adfsInfoPath = "$outputPath\info_adfs.txt"
    Get-AdfsProperties | Format-List * | Out-File $adfsInfoPath
    Log "[adfs] props"
    Get-AdfsGlobalAuthenticationPolicy | Format-List * | Out-File $adfsInfoPath -Append
    Log "[adfs] auth policy"
    Get-AdfsRelyingPartyTrust | Format-List * | Out-File $adfsInfoPath -Append
    Log "[adfs] rps"
    Get-AdfsClaimsProviderTrust -Name 'DigitalPersona STS' | Format-List * | Out-File $adfsInfoPath -Append
    Log "[adfs] idps"
    Get-AdfsAccessControlPolicy | Format-List * | Out-File $adfsInfoPath -Append
    Log "[adfs] acs"
    wevtutil epl "AD FS/Admin" "$outputPath\adfs-admin.evtx" /ow:true
    Log "[adfs] event"
    Log "[adfs] end"
  }
  catch {
    Log "[adfs] error $($_.Exception.Message)"
  }
}

# DP WEB APPS PING

if (Test-Path $wmcPath) {
  if ($stop) {
    Log "[ping] begin"
    @(
      @{ app = "DP Access Mgmt"; path = 'DPWebAUTH/DPWebAuthService.svc/Ping'; file = 'site_dpwebauth.txt' }
      @{ app = "DP STS"; path = 'dppassivests/wsfed/metadata'; file = 'site_dpsts_metadata.txt' }
      @{ app = "DP STS"; path = 'dppassivests/ping'; file = 'site_dpsts_ping.txt' }
      @{ app = "DP Web Enroll"; path = 'dpenrollment'; file = 'site_dpenroll.txt' }            
      @{ app = "DP Web Admin"; path = 'dpadminui'; file = 'site_dpadmin.txt' }
    ) | ForEach-Object { [pscustomobject]$_ } | ForEach-Object {   
      $i = $_   
      Log "[ping] $($i.app) $($i.path)"
      try {
        $u = "$((Get-ItemProperty "HKLM:\Software\DigitalPersona\Applications\ConfigurationTool\$($i.app)\").RootUrl)$($i.path)"
        Log "[ping] $($u)"
        $r = Invoke-WebRequest -Uri $u
        $r.RawContent | Out-File "$($outputPath)\$($i.file)"
      }
      catch {
        $_ | Out-File "$($outputPath)\$($i.file)"
      }
    }
    Log "[ping] end"
  }
}

Log "----END----"