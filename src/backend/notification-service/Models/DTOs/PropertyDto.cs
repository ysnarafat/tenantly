using System.ComponentModel.DataAnnotations;

namespace TenantlyNotificationService.Models.DTOs;

public class PropertyDto
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public string? Address { get; set; }
    public string? City { get; set; }
    public string? PostalCode { get; set; }
    public string? Country { get; set; }
    public int? TotalFloors { get; set; }
    public int? TotalShops { get; set; }
    public bool Active { get; set; }
    public DateTime CreatedAt { get; set; }
    public DateTime UpdatedAt { get; set; }
}

public class CreatePropertyRequest
{
    [Required]
    [MaxLength(100)]
    public string Name { get; set; } = string.Empty;

    [MaxLength(200)]
    public string? Address { get; set; }

    [MaxLength(50)]
    public string? City { get; set; }

    [MaxLength(20)]
    public string? PostalCode { get; set; }

    [MaxLength(50)]
    public string? Country { get; set; } = "Bangladesh";

    [Range(1, 100)]
    public int? TotalFloors { get; set; }

    [Range(1, 1000)]
    public int? TotalShops { get; set; }
}

public class UpdatePropertyRequest
{
    [MaxLength(100)]
    public string? Name { get; set; }

    [MaxLength(200)]
    public string? Address { get; set; }

    [MaxLength(50)]
    public string? City { get; set; }

    [MaxLength(20)]
    public string? PostalCode { get; set; }

    [MaxLength(50)]
    public string? Country { get; set; }

    [Range(1, 100)]
    public int? TotalFloors { get; set; }

    [Range(1, 1000)]
    public int? TotalShops { get; set; }

    public bool? Active { get; set; }
}