using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using TenantlyNotificationService.Models.Enums;
using TenantlyNotificationService.Models.Interfaces;

namespace TenantlyNotificationService.Models.Entities;

[Table("attachments")]
public class Attachment : IAuditableEntity
{
    [Key]
    [Column("id")]
    public int Id { get; set; }

    [Required]
    [MaxLength(200)]
    [Column("file_name")]
    public string FileName { get; set; } = string.Empty;

    [Required]
    [MaxLength(500)]
    [Column("file_path")]
    public string FilePath { get; set; } = string.Empty;

    [Required]
    [MaxLength(100)]
    [Column("content_type")]
    public string ContentType { get; set; } = string.Empty;

    [Column("file_size")]
    public long FileSize { get; set; }

    [Required]
    [Column("attachment_type")]
    public AttachmentType AttachmentType { get; set; }

    [Column("status")]
    public AttachmentStatus Status { get; set; } = AttachmentStatus.Active;

    [MaxLength(500)]
    [Column("description")]
    public string? Description { get; set; }

    [MaxLength(100)]
    [Column("tags")]
    public string? Tags { get; set; }

    // Polymorphic relationships - only one should be set
    [Column("payment_id")]
    public int? PaymentId { get; set; }

    [Column("property_id")]
    public int? PropertyId { get; set; }

    [Column("shop_id")]
    public int? ShopId { get; set; }

    [Column("lease_id")]
    public int? LeaseId { get; set; }

    [Column("tenant_id")]
    public int? TenantId { get; set; }

    [Column("uploaded_by")]
    public int? UploadedBy { get; set; }

    [Column("created_at")]
    public DateTime CreatedAt { get; set; }

    [Column("updated_at")]
    public DateTime UpdatedAt { get; set; }

    // Navigation properties
    [ForeignKey("PaymentId")]
    public virtual Payment? Payment { get; set; }

    [ForeignKey("PropertyId")]
    public virtual Property? Property { get; set; }

    [ForeignKey("ShopId")]
    public virtual Shop? Shop { get; set; }

    [ForeignKey("LeaseId")]
    public virtual Lease? Lease { get; set; }

    [ForeignKey("TenantId")]
    public virtual Tenant? Tenant { get; set; }

    [ForeignKey("UploadedBy")]
    public virtual User? UploadedByUser { get; set; }
}