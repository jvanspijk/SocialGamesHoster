namespace API.Models;

// Maybe needed later to figure out if a user is still logged in and active so that we can block simultaneous access to a single user.
// The user needs to be able to log back in if they disconnect though
public class ActiveSession
{
    public Guid SessionId { get; set; }
    public required string UserName { get; set; }
    public DateTime LastActive { get; set; }
}
