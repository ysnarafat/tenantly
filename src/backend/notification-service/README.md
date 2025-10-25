# Tenantly Notification Service

## Overview

The Tenantly Notification Service is a C# background service responsible for handling automated notifications, reminders, and background processing tasks for the Tenantly shop rental management system. This service uses Entity Framework Core for data access while the database schema is managed by the Go API project.

## Architecture

### Technology Stack
- **.NET 8**: Modern C# runtime with performance improvements
- **Entity Framework Core**: ORM for database operations with PostgreSQL
- **Serilog**: Structured logging framework
- **Background Services**: Hosted services for continuous processing
- **PostgreSQL**: Database (schema managed by Go API)

### Service Responsibilities
- **Notification Processing**: SMS and email delivery to tenants
- **Automated Reminders**: Overdue payment notifications
- **Background Tasks**: Scheduled operations and data processing
- **Audit Logging**: Activity tracking and compliance

### Database Architecture
**Important**: Database migrations and schema management are handled by the Go API project. This service connects to the existing database and uses EF Core for data access only.

## EF Core Implementation

### Migration from Dapper to Entity Framework Core

#### Package Updates
- **Removed**: `Dapper` and `Npgsql` (direct)
- **Added**: `Microsoft.EntityFrameworkCore`, `Microsoft.EntityFrameworkCore.Design`, `Npgsql.EntityFrameworkCore.PostgreSQL`

#### Entity Models
Created comprehensive EF Core entity models in `Models/Entities/`:
- **User.cs** - User entity with proper annotations
- **Property.cs** - Property entity for managing building properties
- **Shop.cs** - Shop entity with navigation properties and Property foreign key
- **Tenant.cs** - Tenant entity with navigation properties  
- **Lease.cs** - Lease entity with foreign key relationships
- **Payment.cs** - Payment entity with foreign key relationships and enum status
- **Attachment.cs** - File attachment entity with polymorphic relationships
- **NotificationQueue.cs** - Updated with EF Core annotations and navigation properties

#### Service Updates
- **DatabaseService.cs** - Converted from Dapper to EF Core:
  - Uses LINQ queries instead of raw SQL
  - Proper async/await patterns
  - Include() for related data loading
  - Scoped service lifetime

#### Worker Updates
- **NotificationWorker.cs** - Updated to use scoped services
- **ReminderWorker.cs** - Updated to use scoped services
- Both now use `IServiceProvider.CreateScope()` for database access

#### Benefits of EF Core Migration
1. **Type Safety**: Compile-time checking of queries and IntelliSense support
2. **Maintainability**: LINQ queries are more readable than raw SQL
3. **Performance**: Query optimization by EF Core and connection pooling
4. **Development Productivity**: Code-first approach and automatic model validation

## Audit Interceptor Implementation

### Automatic Timestamp Management

#### IAuditableEntity Interface
```csharp
public interface IAuditableEntity
{
    DateTime CreatedAt { get; set; }
    DateTime UpdatedAt { get; set; }
}
```

All entities requiring automatic timestamp management implement this interface.

#### AuditInterceptor
The `AuditInterceptor` class extends `SaveChangesInterceptor` and automatically:
- Sets `CreatedAt` and `UpdatedAt` for new entities (EntityState.Added)
- Updates `UpdatedAt` for modified entities (EntityState.Modified)
- Uses UTC timestamps for consistency
- Provides optional logging for debugging

#### Auditable Entities
All entities implement `IAuditableEntity`:
- User, Property, Shop, Tenant, Lease, Payment, Attachment, NotificationQueue

#### Benefits
1. **Database Independence**: Timestamps handled in application code
2. **Consistent UTC Handling**: All timestamps use `DateTime.UtcNow`
3. **Centralized Logic**: All audit timestamp logic in one interceptor
4. **Better Debugging**: Optional logging support for tracking operations
5. **Type Safety**: Interface-based approach ensures compile-time checking
6. **Performance**: More efficient than database triggers

#### Configuration
```csharp
// Register the interceptor as a service
builder.Services.AddScoped<AuditInterceptor>();

builder.Services.AddDbContext<TenantlyDbContext>((serviceProvider, options) =>
    options.UseNpgsql(connectionString)
           .AddInterceptors(serviceProvider.GetRequiredService<AuditInterceptor>()));
```

## Document to Attachment Migration

### Entity Renaming and Enhancement

#### Changes Made
1. **Entity Renaming**
   - **Old**: `Document` entity with `DocumentType` enum
   - **New**: `Attachment` entity with `AttachmentType` enum

2. **Database Schema Changes** (handled by Go API migrations)
   - **Table**: `documents` → `attachments`
   - **Column**: `document_type` → `attachment_type`
   - **Constraint**: `CK_Document_SingleEntity` → `CK_Attachment_SingleEntity`
   - **Indexes**: All indexes updated to reflect new table and column names

3. **Entity Updates**
   All entities with Document navigation properties updated:
   - `Property.Documents` → `Property.Attachments`
   - `Shop.Documents` → `Shop.Attachments`
   - `Tenant.Documents` → `Tenant.Attachments`
   - `Lease.Documents` → `Lease.Attachments`
   - `Payment.Documents` → `Payment.Attachments`

4. **DTO Changes**
   - **Old**: `DocumentDto`, `UploadDocumentRequest`, `UpdateDocumentRequest`
   - **New**: `AttachmentDto`, `UploadAttachmentRequest`, `UpdateAttachmentRequest`

5. **Enum Enhancements**
   The new `AttachmentType` enum includes additional generic types:
   ```csharp
   // New generic attachment types
   Image = 60,
   PDF = 61,
   Spreadsheet = 62,
   TextDocument = 63,
   Other = 99
   ```

#### Benefits of the Change
1. **More Generic Naming**: "Attachment" is more applicable to various file types
2. **Enhanced Type System**: Additional attachment types for better categorization
3. **Improved Clarity**: Better alignment with common file management terminology
4. **Future Extensibility**: Easier to extend with new attachment types

#### Usage Examples
```csharp
// Creating an Attachment
var attachment = new Attachment
{
    FileName = "receipt.pdf",
    AttachmentType = AttachmentType.PaymentReceipt,
    PaymentId = paymentId,
    UploadedBy = userId
};

context.Attachments.Add(attachment);
await context.SaveChangesAsync();

// Querying Attachments
var paymentAttachments = await context.Attachments
    .Where(a => a.PaymentId == paymentId)
    .ToListAsync();

// Navigation Properties
var property = await context.Properties
    .Include(p => p.Attachments)
    .FirstOrDefaultAsync(p => p.Id == propertyId);
```

## Configuration

### Database Connection
```json
{
  "ConnectionStrings": {
    "DefaultConnection": "Host=localhost;Database=tenantly;Username=postgres;Password=password"
  }
}
```

### Notification Settings
```json
{
  "Notifications": {
    "SmsProvider": {
      "ApiUrl": "https://api.sms-provider.com",
      "ApiKey": "your-api-key"
    },
    "EmailProvider": {
      "SmtpHost": "smtp.gmail.com",
      "SmtpPort": 587,
      "Username": "your-email@gmail.com",
      "Password": "your-password"
    }
  }
}
```

### Logging Configuration
```json
{
  "Serilog": {
    "MinimumLevel": "Information",
    "WriteTo": [
      {
        "Name": "Console"
      },
      {
        "Name": "File",
        "Args": {
          "path": "logs/notification-service-.txt",
          "rollingInterval": "Day"
        }
      }
    ]
  }
}
```

## Running the Service

### Prerequisites
1. .NET 8 SDK installed
2. PostgreSQL database running
3. Database schema created by Go API migrations

### Development
```bash
cd src/backend/notification-service
dotnet restore
dotnet run
```

### Production
```bash
dotnet publish -c Release
dotnet TenantlyNotificationService.dll
```

## Database Schema Compatibility

The EF Core models are designed to work with the existing database schema:
- All table names match existing schema (`users`, `shops`, `tenants`, etc.)
- Column names preserved with `[Column]` attributes
- Foreign key relationships maintained and enhanced
- Indexes preserved through Fluent API configuration

## Testing

### Unit Testing
Test the following components:
1. **Audit Interceptor**: Verify timestamps are set correctly
2. **Database Service**: Test LINQ queries and data operations
3. **Notification Workers**: Test background processing logic
4. **Attachment Operations**: Verify CRUD operations work correctly

### Integration Testing
1. **Database Connectivity**: Ensure EF Core connects to PostgreSQL
2. **Entity Relationships**: Test navigation properties function correctly
3. **Background Services**: Verify workers process notifications correctly

## Troubleshooting

### Common Issues

#### Timestamps Not Being Set
- **Cause**: Entity doesn't implement `IAuditableEntity`
- **Solution**: Ensure entity implements the interface

#### Interceptor Not Running
- **Cause**: Interceptor not registered in DI container
- **Solution**: Verify interceptor registration in `Program.cs`

#### Database Connection Issues
- **Cause**: Incorrect connection string or database not available
- **Solution**: Check connection string and ensure PostgreSQL is running

#### Migration Errors
- **Cause**: Attempting to run EF Core migrations
- **Solution**: Remember that migrations are handled by Go API, not this service

### Debug Logging
Enable debug logging for specific components:
```json
{
  "Logging": {
    "LogLevel": {
      "TenantlyNotificationService.Data.Interceptors.AuditInterceptor": "Debug",
      "TenantlyNotificationService.Services.DatabaseService": "Debug"
    }
  }
}
```

## Future Considerations

### Attachment Management
1. **File Storage**: Consider implementing cloud storage integration
2. **Versioning**: Add support for attachment versioning
3. **Metadata**: Enhance metadata support for different file types
4. **Security**: Implement role-based access control for attachments
5. **Thumbnails**: Add thumbnail generation for image attachments

### Performance Optimization
1. **Caching**: Implement Redis caching for frequently accessed data
2. **Query Optimization**: Add query performance monitoring
3. **Background Processing**: Optimize worker service performance
4. **Connection Pooling**: Fine-tune EF Core connection pooling

### Monitoring and Observability
1. **Health Checks**: Add comprehensive health check endpoints
2. **Metrics**: Implement application metrics collection
3. **Distributed Tracing**: Add OpenTelemetry support
4. **Alerting**: Set up monitoring and alerting for critical operations