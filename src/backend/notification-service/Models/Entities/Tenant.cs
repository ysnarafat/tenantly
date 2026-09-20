using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using TenantlyNotificationService.Models.Interfaces;

namespace TenantlyNotificationService.Models.Entities;

[Table("tenants")]
public class Tenant : IAuditableEntity
{
    [Key]
    [Column("id")]
    public int Id { get; set; }

    [Required]
    [MaxLength(100)]
    [Column("name")]
    public string Name { get; set; } = string.Empty;

    [MaxLength(20)]
    [Column("phone_number")]
    public string? PhoneNumber { get; set; }

    [MaxLength(100)]
    [Column("email")]
    public string? Email { get; set; }

    [Column("address")]
    public string? Address { get; set; }

    [Column("active")]
    public bool Active { get; set; } = true;

    [Column("created_at")]
    public DateTime CreatedAt { get; set; }

    [Column("updated_at")]
    public DateTime UpdatedAt { get; set; }

    // Navigation properties
    public virtual ICollection<Lease> Leases { get; set; } = new List<Lease>();
    public virtual ICollection<Payment> Payments { get; set; } = new List<Payment>();
    public virtual ICollection<NotificationQueue> Notifications { get; set; } = new List<NotificationQueue>();
}