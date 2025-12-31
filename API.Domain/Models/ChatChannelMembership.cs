using System.Threading.Channels;

namespace API.Domain.Models;

public class ChatChannelMembership
{
    public required Player Player { get; set; }
    public required int PlayerId { get; set; }
    public required ChatChannel Channel { get; set; }
    public required int ChannelId { get; set; }
}
