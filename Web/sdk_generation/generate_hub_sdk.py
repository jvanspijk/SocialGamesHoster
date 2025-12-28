import pathlib
from pathlib import Path
import re
import os
import textwrap

def map_csharp_to_ts(csharp_type: str) -> str:
    csharp_type = csharp_type.strip()
    
    is_nullable = False
    if csharp_type.endswith('?'):
        is_nullable = True
        csharp_type = csharp_type[:-1]

    list_match = re.search(r'(?:List|IEnumerable|ICollection|IList)<(.*)>', csharp_type)
    if list_match:
        inner_type = map_csharp_to_ts(list_match.group(1))
        ts_type = f"{inner_type}[]"
    else:
        mapping = {
            "int": "number",
            "long": "number",
            "float": "number",
            "double": "number",
            "decimal": "number",
            "bool": "boolean",
            "string": "string",
            "Guid": "string",
            "DateTime": "string",
            "DateTimeOffset": "string",
            "object": "any",
            "void": "void"
        }
        ts_type = mapping.get(csharp_type, csharp_type)

    return f"{ts_type} | null" if is_nullable else ts_type

def get_url_from_hub(hub: Hub) -> str:
    return f"api/{hub.feature_name}/hub"

def create_hub_ts_file(hub: Hub) -> str:
    indent = "    "
    
    event_types = []
    state_inits = []
    listeners = []

    for method in hub.methods:
        method.add_member("ts", "number")
        
        event_types.append(f"{indent}{method.name}: {{ {method.to_ts_params()} }};")
        state_inits.append(f"{indent}{indent}{method.name}: null,")

        listener = textwrap.dedent(f"""\
        this.connection.on("{method.name}", ({method.get_listener_args()}) => {{
                this.#state.{method.name} = {{ {method.get_assignment_body()} }};
            }});""")
            
        listeners.append(textwrap.indent(listener, indent * 2))

    events_str = "\n".join(event_types)
    state_str = "\n".join(state_inits)
    listeners_str = "\n".join(listeners)
    feature_lower = hub.feature_name[0].lower() + hub.feature_name[1:]


    template = f"""\
    import * as signalR from '@microsoft/signalr';
    import {{ browser }} from '$app/environment';
    import {{ untrack }} from 'svelte';

    export type {hub.feature_name}HubEvents = {{
    {events_str}
    }};

    class {hub.feature_name}Hub {{
        private connection: signalR.HubConnection | null = null;
        
        #state = $state<{{ [K in keyof {hub.feature_name}HubEvents]: {hub.feature_name}HubEvents[K] | null }}>({{
    {state_str}
        }});

        #connectionState = $state<"Disconnected" | "Connected" | "Reconnecting" | "Faulted">("Disconnected");

        constructor(url: string) {{
            if (!browser) return;

            this.connection = new signalR.HubConnectionBuilder()
                .withUrl(url)
                .withAutomaticReconnect()
                .configureLogging(signalR.LogLevel.Information)
                .build();

            this.registerListeners();
            this.start();
        }}

        private async start() {{
            try {{
                await this.connection?.start();
                this.#connectionState = "Connected";
            }} catch (err) {{
                console.error("SignalR Start Error: ", err);
                this.#connectionState = "Faulted";
            }}
        }}

        private registerListeners() {{
            if (!this.connection) return;

    {listeners_str}

            this.connection.onreconnecting(() => this.#connectionState = "Reconnecting");
            this.connection.onreconnected(() => this.#connectionState = "Connected");
            this.connection.onclose(() => this.#connectionState = "Disconnected");
        }}

        get events() {{ return this.#state; }}
        get status() {{ return this.#connectionState; }}

        onEvent<K extends keyof {hub.feature_name}HubEvents>(
            key: K, 
            callback: (payload: {hub.feature_name}HubEvents[K]) => void
        ) {{
            $effect.pre(() => {{
                const value = this.#state[key];
                if (value) {{
                    untrack(() => callback(value));
                    this.#state[key] = null;
                }}
            }});
        }}

        async disconnect() {{
            if (this.connection) {{
                await this.connection.stop();
            }}
        }}
    }}

    const apiBase = browser ? `${{window.location.protocol}}//${{window.location.hostname}}:9090` : '';
    export const {feature_lower}Hub = new {hub.feature_name}Hub(`${{apiBase}}/api/{feature_lower}/hub`);
    """

    return textwrap.dedent(template)
    


def create_sdk_hubs(hubs: List[Hub], output_base_path: Path):
    for hub in hubs:
        if not hub.methods:
            print(f"Warning! No methods found in {hub.path}")
            continue
        file_content = create_hub_ts_file(hub)
    
        filename = f"{hub.feature_name}Hub.svelte.ts"
        
        output_file = output_base_path / hub.feature_name / filename

        with open(output_file, "w", encoding="utf-8") as f:
            f.write(file_content)