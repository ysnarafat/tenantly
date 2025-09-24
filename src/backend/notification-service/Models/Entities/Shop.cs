using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using TenantlyNotificationService.Models.Interfaces;

namespace TenantlyNotificationService.Models.Entities;

[Table("shops")]
public class Shop : IAuditableEntity
{
    [Key]
    [Column("id")]
    public int Id { get; set; }

    [Required]
    [MaxLength(100)]
    [Column("name")]
    public string Name { get; set; } = string.Empty;

    [Required]
    [MaxLength(20)]
    [Column("shop_number")]
    public string ShopNumber { get; set; } = string.Empty;

    [MaxLength(10)]
    [Column("floor")]
    public string? Floor { get; set; }

    [MaxLength(50)]
    [Column("section")]
    public string? Section { get; set; }

    [Required]
    [Column("monthly_rent", TypeName = "decimal(10,2)")]
    public decimal MonthlyRent { get; set; }

    [Column("active")]
    public bool Active { get; set; } = true;

    [Column("property_id")]
    public int PropertyId { get; set; } = 1;

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
    public virtual ICollection<Attachment> Attachments { get; set; } = new List<Attachment>();
}