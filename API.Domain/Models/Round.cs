using System.Text.Json.Serialization;

namespace API.Domain.Models;

public class Round
{
    public int Id { get; set; }
    public DateTimeOffset? StartedTime {  get; set; }
    public DateTimeOffset? FinishedTime { get; set; }
    [JsonIgnore]
    public GameSession? GameSession { get; set; }
    public int GameId { get; set; }
   
}
