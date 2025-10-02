namespace API.DataAccess;

public interface IRepository<T>
{
    // Create
    Task<T> CreateAsync(T entity);
    // Read
    Task<T?> GetByIdAsync(int id);
    Task<List<T>> GetAllAsync();
    // Update
    Task UpdateAsync(T entity);
    // Delete
    Task DeleteAsync(T entity);
}
