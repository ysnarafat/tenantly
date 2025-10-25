using System.ComponentModel.DataAnnotations;
using TenantlyNotificationService.Models.Enums;

namespace TenantlyNotificationService.Models.DTOs;

public class AttachmentDto
{
    public int Id { get; set; }
    public string FileName { get; set; } = string.Empty;
    public string FilePath { get; set; } = string.Empty;
    public string ContentType { get; set; } = string.Empty;
    public long FileSize { get; set; }
    public AttachmentType AttachmentType { get; set; }
    public AttachmentStatus Status { get; set; }
    public string? Description { get; set; }
    public string? Tags { get; set; }
    public int? PaymentId { get; set; }
    public int? PropertyId { get; set; }
    public int? ShopId { get; set; }
    public int? LeaseId { get; set; }
    public int? TenantId { get; set; }
    public int? UploadedBy { get; set; }
    public string? UploadedByName { get; set; }
    public DateTime CreatedAt { get; set; }
    public DateTime UpdatedAt { get; set; }
}

public class UploadAttachmentRequest
{
    [Required]
    public IFormFile File { get; set; } = null!;

    [Required]
    public AttachmentType AttachmentType { get; set; }

    [MaxLength(500)]
    public string? Description { get; set; }

    [MaxLength(100)]
    public string? Tags { get; set; }

    // Only one of these should be provided
    public int? PaymentId { get; set; }
    public int? PropertyId { get; set; }
    public int? ShopId { get; set; }
    public int? LeaseId { get; set; }
    public int? TenantId { get; set; }
}

public class UpdateAttachmentRequest
{
    [MaxLength(500)]
    public string? Description { get; set; }

    [MaxLength(100)]
    public string? Tags { get; set; }

    public AttachmentStatus? Status { get; set; }
}

public class AttachmentSearchRequest
{
    public AttachmentType? AttachmentType { get; set; }
    public AttachmentStatus? Status { get; set; }
    public int? PaymentId { get; set; }
    public int? PropertyId { get; set; }
    public int? ShopId { get; set; }
    public int? LeaseId { get; set; }
    public int? TenantId { get; set; }
    public int? UploadedBy { get; set; }
    public DateTime? FromDate { get; set; }
    public DateTime? ToDate { get; set; }
    public string? SearchTerm { get; set; }
    public int Page { get; set; } = 1;
    public int PageSize { get; set; } = 20;
}

public class AttachmentSummaryDto
{
    public string EntityType { get; set; } = string.Empty;
    public int EntityId { get; set; }
    public string EntityName { get; set; } = string.Empty;
    public int TotalAttachments { get; set; }
    public Dictionary<AttachmentType, int> AttachmentsByType { get; set; } = new();
    public DateTime? LastUploadDate { get; set; }
}