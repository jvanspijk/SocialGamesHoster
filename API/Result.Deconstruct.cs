using API.Validation;

namespace API;

public readonly partial record struct Result<T>
{    
    public void Deconstruct(out T? value, out APIError? error)
    {
        value = Value; 
        error = Error;
    }
}
