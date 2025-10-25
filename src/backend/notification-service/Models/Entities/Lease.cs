using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using TenantlyNotificationService.Models.Interfaces;

namespace TenantlyNotificationService.Models.Entities;

[Table("leases")]
public class Lease : IAuditableEntity
{
    [Key]
    [Column("id")]
    public int Id { get; set; }

    [Required]
    [Column("shop_id")]
    public int ShopId { get; set; }

    [Required]
    [Column("tenant_id")]
    public int TenantId { get; set; }

    [Required]
    [Column("start_date")]
    public DateOnly StartDate { get; set; }

    [Required]
    [Column("duration_months")]
    public int DurationMonths { get; set; }

    [Required]
    [Column("monthly_rent", TypeName = "decimal(10,2)")]
    public decimal MonthlyRent { get; set; }

    [Column("security_deposit", TypeName = "decimal(10,2)")]
    public decimal? SecurityDeposit { get; set; }

    [Column("active")]
    public bool Active { get; set; } = true;

    [Column("created_at")]
    public DateTime CreatedAt { get; set; }

    [Column("updated_at")]
    public DateTime UpdatedAt { get; set; }

    // Navigation properties
    [ForeignKey("ShopId")]
    public virtual Shop Shop { get; set; } = null!;

    [ForeignKey("TenantId")]
    public virtual Tenant Tenant { get; set; } = null!;
    
    public virtual ICollection<Attachment> Attachments { get; set; } = new List<Attachment>();
}