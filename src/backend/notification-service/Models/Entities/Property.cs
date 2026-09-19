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
    [MaxLength(200)]
    [Column("property_name")]
    public string PropertyName { get; set; } = string.Empty;

    [Required]
    [MaxLength(50)]
    [Column("property_code")]
    public string PropertyCode { get; set; } = string.Empty;

    [Required]
    [Column("address")]
    public string Address { get; set; } = string.Empty;

    [MaxLength(100)]
    [Column("city")]
    public string? City { get; set; }

    [MaxLength(20)]
    [Column("postal_code")]
    public string? PostalCode { get; set; }

    [Column("active")]
    public bool Active { get; set; } = true;

    [Column("created_at")]
    public DateTime CreatedAt { get; set; }

    [Column("updated_at")]
    public DateTime UpdatedAt { get; set; }

    // Navigation properties
    public virtual ICollection<Unit> Units { get; set; } = new List<Unit>();
}
