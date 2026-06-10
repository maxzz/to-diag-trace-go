import { UISymbolDefsInject } from "./symbols-inject";
import { DefAppTypes } from "./app";
import { DefAllOtherTypes } from "./all-other";

export * from "./app";
export * from "./all-other";

export function UISymbolDefs() {
    return (
        <UISymbolDefsInject>
            {DefAppTypes()}
            {DefAllOtherTypes()}
        </UISymbolDefsInject>
    );
}
