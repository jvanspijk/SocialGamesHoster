import pathlib
from pathlib import Path
from dataclasses import dataclass
from typing import Dict, List
import re
import urllib.request
import json
import shutil
from generate_hub_sdk import create_sdk_hubs
from generate_api_sdk import create_api_sdk


def clear_output_dir(output_base_path: Path):
    dont_delete = ["api.ts", "config.json"]

    if not output_base_path.exists():
        raise Exception("Output path not found at {}".format(output_base_path))
        

    for item in output_base_path.iterdir():
        if item.name in dont_delete:
            continue
        
        if item.is_dir():
            shutil.rmtree(item)
        else:
            item.unlink()

class Endpoint:
    def __init__(self, path: Path, open_api: dict):
        self.path = path
        self.endpoint_name = path.stem
        self.feature_name = path.parent.parent.name
        self.url = None
        self.method = None
        self.returns_array = False
        self.set_url_and_method(open_api)
        self.placeholders = re.findall(r'\{([^}]+)\}', self.url)

    def set_url_and_method(self, open_api):
        for path in open_api["paths"]:
            for method in open_api["paths"][path]:
                try:                    
                    if(open_api["paths"][path][method]["operationId"] == self.endpoint_name):
                        self.url = path
                        self.method = method
                        try:
                            if open_api["paths"][path][method]["responses"]["200"]["content"]["application/json"]["schema"]["type"] == "array":
                                self.returns_array = True
                        except KeyError:
                            self.returns_array = False
                        return
                except KeyError:
                    continue

        print(f"url of {self.endpoint_name} not found!")
        clear_output_dir()
        exit()

class Common:
    def __init__(self, path: Path):
        self.path = path
        self.file_name = path.stem
        self.feature_name = path.parent.parent.name

class Method:
    def __init__(self, name):
        self.name = name
        self.params = {}

    def to_ts_params(self) -> str:
        return " ".join([f"{name}: {type};" for name, type in self.params.items()])

    def get_listener_args(self) -> str:
        args = [f"{n}: {t}" for n, t in self.params.items() if n != "ts"]
        return ", ".join(args)

    def get_assignment_body(self) -> str:
        parts = ["ts: Date.now()" if n == "ts" else n for n in self.params.keys()]
        return ", ".join(parts)

    def __repr__(self) -> str:
        param_list = [f"{param_type} {param_name}" for param_name, param_type in self.params.items()]            
        params_str = ", ".join(param_list)    
        return f"{self.name}({params_str})"

    def add_member(self, name: str, type: str):
        self.params[name] = map_csharp_to_ts(type)

class Hub:
    def __init__(self, path: Path):
        self.path = path
        self.file_name = path.stem
        self.feature_name = path.parent.parent.name
        self.url = f"api/{self.feature_name}/hub"
        self.methods = []

    def __repr__(self) -> str:
        repr = f"{self.feature_name}Hub ({self.url}):\n"        
        for method in self.methods:
            repr += f"    {method}\n"
        return repr

    def add_method(self, method: Method):
        self.methods.append(method)

class TypescriptClass:
    def __init__(self, name: str, members: Dict[str, str]):
        self.name = name
        self.members = members

    def to_string(self, indent_level: int = 0) -> str:
        indent = "    " * indent_level
        members_str = ""
        for name, type_ in self.members.items():
            members_str += f"{indent}    readonly {name}: {type_};\n"
        return f"{indent}export type {self.name} = {{\n{members_str}{indent}}}"

def get_hub_from_file(path: Path) -> Hub:
    hub = Hub(path)
    with open(path, "r") as f:
        content = f.read()
    
    interface_pattern = r'interface\s+I\w+Hub\s*\{([\s\S]*?)\}'
    interface_match = re.search(interface_pattern, content)

    if interface_match:
        interface_body = interface_match.group(1)
        method_pattern = r'Task\s+(\w+)\((.*?)\);'
        methods = re.finditer(method_pattern, interface_body)

        for m_match in methods:
            method_name = m_match.group(1)
            raw_params = m_match.group(2)
            
            method_obj = Method(method_name)

            if raw_params.strip():
                param_pairs = [p.strip() for p in raw_params.split(',')]
                for pair in param_pairs:
                    parts = pair.split()
                    if len(parts) == 2:
                        param_type, param_name = parts
                        method_obj.add_member(param_name, param_type)
            
            hub.add_method(method_obj)
    return hub


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

def GetTypeFromEndpoint(endpoint: Endpoint, object_name: str) -> TypescriptClass | None:
    ts_class_name = f"{endpoint.endpoint_name}{object_name}"

    with open(endpoint.path, "r") as f:
        content = f.read()

    pattern = r'public\s+(?:readonly\s+)?record\s+(?:struct\s+)?' + object_name + r'\s*\((.*?)\)'
    
    match = re.search(pattern, content, re.DOTALL)

    members = {placeholder: "string" for placeholder in endpoint.placeholders} if object_name == "Request" else {}
    if match:
        params_content = match.group(1).strip()
        parts = re.split(r',\s*(?![^<>]*>)', params_content)
        
        for part in parts:
            part = part.strip()
            if not part: continue
            
            type_str, name_str = part.rsplit(' ', 1)
            # convert to camel case
            name_str = name_str[0].lower() + name_str[1:]
            
            members[name_str] = map_csharp_to_ts(type_str)
    return TypescriptClass(ts_class_name, members) if members else None

def GetStructsFromCommon(common: Common) -> List[TypescriptClass]:
    with open(common.path, "r") as f:
        content = f.read()
    pattern = r'public\s+(?:readonly\s+)?record\s+(?:struct\s+)?(\w+)\s*\((.*?)\)'
    results = []
    for match in re.finditer(pattern, content, re.DOTALL):
        struct_name = match.group(1)
        params_content = match.group(2).strip()
        
        parts = re.split(r',\s*(?![^<>]*>)', params_content)
        
        members = {}
        for part in parts:
            part = part.strip()
            if not part: continue
            
            # Extract Type and Name
            type_str, name_str = part.rsplit(' ', 1)
            
            ts_type = map_csharp_to_ts(type_str)
            # convert to camel case
            name_str = name_str[0].lower() + name_str[1:]
            members[name_str] = ts_type

        results.append(TypescriptClass(struct_name, members))
        
    return results

def load_openapi_json(open_api_url: str):
    with urllib.request.urlopen(open_api_url) as response:
        data = response.read().decode('utf-8')
        return json.loads(data)


def create_sdk_common_files(common_files: List[Common], output_base_path: Path, feature_common_types_dict: Dict[str, set]):
    for common in common_files:
        feature_dir = output_base_path / common.feature_name
        feature_dir.mkdir(parents=True, exist_ok=True)
        output_path = feature_dir / "Common.ts"

        ts_structs = GetStructsFromCommon(common)
        if not ts_structs or ts_structs == []:
            print(f"Warning! No structs found in {common.path}")
            continue

        with open(output_path, "w") as f:
            for ts_struct in ts_structs:
                f.write(ts_struct.to_string() + ";\n\n")
                feature_common_types_dict[common.feature_name].add(ts_struct.name)

def create_sdk_endpoints(endpoints: List[Endpoint], output_base_path: Path, feature_common_types_dict: Dict[str, set]):
    for endpoint in endpoints:
        feature_dir = output_base_path / endpoint.feature_name
        feature_dir.mkdir(parents=True, exist_ok=True)
        output_path = feature_dir / f"{endpoint.endpoint_name}.ts"

        ts_request = GetTypeFromEndpoint(endpoint, "Request")
        ts_response = GetTypeFromEndpoint(endpoint, "Response")
        endpoint_factory_import_header = "import { createEndpoint } from \"../api\";\n"
        with(open(output_path, "w")) as f:
            f.write(endpoint_factory_import_header)

        if not ts_request and not ts_response:
            continue

        used_common_types = set()
        all_members = []
        if ts_request: all_members.extend(ts_request.members.values())
        if ts_response: all_members.extend(ts_response.members.values())

        for member_type in all_members:
            for common_type in feature_common_types_dict[endpoint.feature_name]:
                if common_type in member_type:
                    used_common_types.add(common_type)

        with(open(output_path, "a")) as f:
            if used_common_types:
                import_names = ", ".join(sorted(used_common_types))
                f.write(f"import type {{ {import_names} }} from './Common';\n\n")
            if ts_request:
                f.write(ts_request.to_string() + ";" + "\n")
            else:
                f.write(f"export type {endpoint.endpoint_name}Request = void;\n")
            if ts_response:
                f.write(ts_response.to_string() + ";" + "\n")
            input_type = ts_request.name if ts_request else (endpoint.endpoint_name + "Request")
            output_type = ts_response.name if ts_response else 'void'
            if endpoint.returns_array:
                output_type = f"{output_type}[]"
            f.write(f"export const {endpoint.endpoint_name} = createEndpoint<{input_type}, {output_type}>")
            f.write(f"('{endpoint.url}', '{endpoint.method.upper()}');\n")

if __name__ == '__main__':
    curr_path = pathlib.Path.cwd()
    features_path = curr_path.parent / "API" / "Features"     
    output_base_path = curr_path / "src" / "lib" / "client"
    open_api_url = "http://localhost:9090/openapi/v1.json"

    feature_structure = {"Endpoints": [], "Common": [], "Hubs": []}

    try:
        clear_output_dir(output_base_path)
        api_json = load_openapi_json(open_api_url)

        if not features_path.exists():
            raise Exception("Features path not found at {}".format(features_path))
            exit()
        
        all_features = set()
        for feature_dir in (d for d in features_path.iterdir() if d.is_dir()):
            all_features.add(feature_dir.name)
            for sub_dir in feature_dir.iterdir():
                if not sub_dir.is_dir():
                    continue
                    
                if sub_dir.name == "Endpoints":
                    for file in sub_dir.glob("*.cs"):
                        obj = Endpoint(file, api_json)
                        feature_structure["Endpoints"].append(obj)
                        
                elif sub_dir.name == "Common":
                    for file in sub_dir.glob("*.cs"):
                        obj = Common(file)
                        feature_structure["Common"].append(obj)

                elif sub_dir.name == "Hubs":
                    for file in sub_dir.glob("*.cs"):
                        obj = get_hub_from_file(file)
                        feature_structure["Hubs"].append(obj)

        feature_common_types_dict = {feature: set() for feature in all_features}
        for feature in all_features:
            feature_output_dir = output_base_path / feature
            feature_output_dir.mkdir(parents=True, exist_ok=True)

        create_sdk_common_files(feature_structure["Common"], output_base_path, feature_common_types_dict)
        create_sdk_endpoints(feature_structure["Endpoints"], output_base_path, feature_common_types_dict)
        create_sdk_hubs(feature_structure["Hubs"], output_base_path)
        create_api_sdk(output_base_path)

    except Exception as e:
        print(f"Error: {e}")
        clear_output_dir(output_base_path)

