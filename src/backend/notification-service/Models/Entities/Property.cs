using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using TenantlyNotificationService.Models.Interfaces;

namespace TenantlyNotificationService.Models.Entities;

[Table("properties")]
public class Property : IAuditableEntity
{
    [Key]
    [Column("id")]
    public int Id { get; set; }

    [Required]
    [MaxLength(100)]
    [Column("name")]
    public string Name { get; set; } = string.Empty;

    [MaxLength(200)]
    [Column("address")]
    public string? Address { get; set; }

    [MaxLength(50)]
    [Column("city")]
    public string? City { get; set; }

    [MaxLength(20)]
    [Column("postal_code")]
    public string? PostalCode { get; set; }

    [MaxLength(50)]
    [Column("country")]
    public string? Country { get; set; } = "Bangladesh";

    [Column("total_floors")]
    public int? TotalFloors { get; set; }

    [Column("total_shops")]
    public int? TotalShops { get; set; }

    [Column("active")]
    public bool Active { get; set; } = true;

    [Column("created_at")]
    public DateTime CreatedAt { get; set; }

    [Column("updated_at")]
    public DateTime UpdatedAt { get; set; }

    // Navigation properties
    public virtual ICollection<Shop> Shops { get; set; } = new List<Shop>();
    public virtual ICollection<Attachment> Attachments { get; set; } = new List<Attachment>();
}