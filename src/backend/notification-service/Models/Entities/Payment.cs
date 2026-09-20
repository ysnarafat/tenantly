using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using TenantlyNotificationService.Models.Enums;
using TenantlyNotificationService.Models.Interfaces;

namespace TenantlyNotificationService.Models.Entities;

[Table("payments")]
public class Payment : IAuditableEntity
{
    [Key]
    [Column("id")]
    public int Id { get; set; }

    [Required]
    [Column("unit_id")]
    public int UnitId { get; set; }

    [Required]
    [Column("tenant_id")]
    public int TenantId { get; set; }

    [Required]
    [Range(1, 12)]
    [Column("month")]
    public int Month { get; set; }

    [Required]
    [Column("year")]
    public int Year { get; set; }

    [Required]
    [Column("amount_due", TypeName = "decimal(10,2)")]
    public decimal AmountDue { get; set; }

    [Column("amount_paid", TypeName = "decimal(10,2)")]
    public decimal AmountPaid { get; set; } = 0;

    [Column("status")]
    public PaymentStatus Status { get; set; } = PaymentStatus.Due;

    [MaxLength(50)]
    [Column("payment_method")]
    public string? PaymentMethod { get; set; }

    [Column("payment_date")]
    public DateOnly? PaymentDate { get; set; }

    [Column("notes")]
    public string? Notes { get; set; }

    [MaxLength(50)]
    [Column("receipt_number")]
    public string? ReceiptNumber { get; set; }

    [Column("due_date")]
    public DateOnly? DueDate { get; set; }

    [Column("created_at")]
    public DateTime CreatedAt { get; set; }

    [Column("updated_at")]
    public DateTime UpdatedAt { get; set; }

    // Navigation properties
    [ForeignKey("UnitId")]
    public virtual Unit Unit { get; set; } = null!;

    [ForeignKey("TenantId")]
    public virtual Tenant Tenant { get; set; } = null!;
}