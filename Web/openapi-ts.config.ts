import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: 'http://localhost:9090/openapi/v1.json',
  output: 'src/lib/client',
  useUnionTypes: true,
});