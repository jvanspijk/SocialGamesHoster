using Microsoft.AspNetCore.OpenApi;
using Microsoft.OpenApi.Models;

namespace API;

/// <summary>
/// Custom schema transformer to avoid name collisions for nested classes like Request and Response.
/// Might be fixed in .NET 10 as these kind of name collisions are common in CQRS and vertical slices.
/// </summary>
public class NestedClassSchemaTransformer : IOpenApiSchemaTransformer
{
    public Task TransformAsync(OpenApiSchema schema, OpenApiSchemaTransformerContext context, CancellationToken cancellationToken)
    {
        var type = context.JsonTypeInfo.Type;

        if (!type.IsNested)
            return Task.CompletedTask;

        if (type.Name.Equals("Request", StringComparison.Ordinal))
        {
            schema.Description = $"Request for {type.DeclaringType?.Name}";
        }
        else if (type.Name.Equals("Response", StringComparison.Ordinal))
        {
            schema.Description = $"Response for {type.DeclaringType?.Name}";
        }

        const string schemaId = "x-schema-id";
        schema.Annotations[schemaId] = $"{type.DeclaringType?.Name}{type.Name}";

        return Task.CompletedTask;
    }  

}
