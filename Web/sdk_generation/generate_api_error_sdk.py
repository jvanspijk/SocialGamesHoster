from pathlib import Path

error_response_type_str = r"""
export type ApiError = {
	status: number;
	title: string;
	errors?: Record<string, string[]>;
	detail?: string;
};
"""

def create_api_error_sdk(output_base_path: Path):
    output_file = output_base_path / "ApiError.ts"
    with open(output_file, "w") as f:
        f.write("// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION\n")
        f.write(error_response_type_str)
        f.close()