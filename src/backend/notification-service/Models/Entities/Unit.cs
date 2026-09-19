using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using TenantlyNotificationService.Models.Interfaces;

namespace TenantlyNotificationService.Models.Entities;

[Table("units")]
public class Unit : IAuditableEntity
{
    [Key]
    [Column("id")]
    public int Id { get; set; }

    [Required]
    [MaxLength(50)]
    [Column("unit_number")]
    public string UnitNumber { get; set; } = string.Empty;

    [MaxLength(100)]
    [Column("unit_name")]
    public string? UnitName { get; set; }

    [Column("property_id")]
    public int PropertyId { get; set; }

    [Column("active")]
    public bool Active { get; set; } = true;

    [Column("created_at")]
    public DateTime CreatedAt { get; set; }

    [Column("updated_at")]
    public DateTime UpdatedAt { get; set; }

    // Navigation properties
    [ForeignKey("PropertyId")]
    public virtual Property Property { get; set; } = null!;

    public virtual ICollection<Lease> Leases { get; set; } = new List<Lease>();
    public virtual ICollection<Payment> Payments { get; set; } = new List<Payment>();
    public virtual ICollection<NotificationQueue> Notifications { get; set; } = new List<NotificationQueue>();
}
