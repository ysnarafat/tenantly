using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using TenantlyNotificationService.Models.Entities;
using TenantlyNotificationService.Models.Interfaces;

namespace TenantlyNotificationService.Models;

[Table("notification_queue")]
public class NotificationQueue : IAuditableEntity
{
    [Key]
    [Column("id")]
    public int Id { get; set; }

    [Required]
    [Column("tenant_id")]
    public int TenantId { get; set; }

    [Required]
    [Column("unit_id")]
    public int UnitId { get; set; }

    [Required]
    [Column("message")]
    public string Message { get; set; } = string.Empty;

    [Required]
    [MaxLength(20)]
    [Column("notification_type")]
    public string NotificationType { get; set; } = string.Empty;

    [Required]
    [MaxLength(100)]
    [Column("recipient")]
    public string Recipient { get; set; } = string.Empty;

    [MaxLength(20)]
    [Column("status")]
    public string Status { get; set; } = "Pending";

    [Column("retry_count")]
    public int RetryCount { get; set; } = 0;

    [Column("error_message")]
    public string? ErrorMessage { get; set; }

    [Column("created_at")]
    public DateTime CreatedAt { get; set; }

    [Column("updated_at")]
    public DateTime UpdatedAt { get; set; }

    // Navigation properties
    [ForeignKey("TenantId")]
    public virtual Tenant Tenant { get; set; } = null!;

    [ForeignKey("UnitId")]
    public virtual Unit Unit { get; set; } = null!;
}

public enum NotificationType
{
    SMS,
    Email,
    Reminder
}

public enum NotificationStatus
{
    Pending,
    Sent,
    Failed
}