# EF Core Migration Guide

## Overview
This document outlines the migration from Dapper with raw SQL to Entity Framework Core for the Tenantly Notification Service.

## Changes Made

### 1. Package Updates
- Removed: `Dapper` and `Npgsql` (direct)
- Added: `Microsoft.EntityFrameworkCore`, `Microsoft.EntityFrameworkCore.Design`, `Npgsql.EntityFrameworkCore.PostgreSQL`

### 2. New Entity Models
Created EF Core entity models in `Models/Entities/`:
- `User.cs` - User entity with proper annotations
- `Property.cs` - Property entity for managing building properties
- `Shop.cs` - Shop entity with navigation properties and Property foreign key
- `Tenant.cs` - Tenant entity with navigation properties  
- `Lease.cs` - Lease entity with foreign key relationships
- `Payment.cs` - Payment entity with foreign key relationships and enum status

### 3. Updated Models
- `NotificationQueue.cs` - Updated with EF Core annotations and navigation properties
- `Payment.cs` - Updated Status property to use PaymentStatus enum instead of string
- Added `PaymentStatus` enum with values: Due, Paid, Partial
- Added `PropertyDto.cs` for Property data transfer objects

### 4. DbContext
- `Data/TenantlyDbContext.cs` - Main EF Core context with:
  - Entity configurations
  - Index definitions
  - Foreign key relationships
  - Default value configurations

### 5. Service Updates
- `DatabaseService.cs` - Converted from Dapper to EF Core:
  - Uses LINQ queries instead of raw SQL
  - Proper async/await patterns
  - Include() for related data loading
  - Scoped service lifetime

### 6. Worker Updates
- `NotificationWorker.cs` - Updated to use scoped services
- `ReminderWorker.cs` - Updated to use scoped services
- Both now use `IServiceProvider.CreateScope()` for database access

### 7. Configuration Updates
- `Program.cs` - Added EF Core registration and database migration
- `appsettings.json` - Added standard ConnectionStrings section
- `DatabaseMigrationService.cs` - Ensures database is up to date on startup

## Benefits of EF Core Migration

### 1. Type Safety
- Compile-time checking of queries
- IntelliSense support for entity properties
- Reduced runtime errors from SQL typos

### 2. Maintainability
- LINQ queries are more readable than raw SQL
- Automatic change tracking
- Built-in migration system

### 3. Performance
- Query optimization by EF Core
- Connection pooling
- Lazy loading capabilities

### 4. Development Productivity
- Code-first approach
- Automatic model validation
- Rich navigation properties

## Database Schema Compatibility

The EF Core models are designed to work with the existing database schema:
- All table names match existing schema (`users`, `shops`, `tenants`, etc.)
- Column names preserved with `[Column]` attributes
- Foreign key relationships maintained and enhanced
- Indexes preserved through Fluent API configuration

### New Schema Additions
- `properties` table added to properly handle the PropertyId foreign key in shops
- Default property record inserted to maintain existing shop references
- Foreign key constraint added between shops and properties tables
- PaymentStatus enum configured to store as string in database for compatibility

## Migration Commands

To generate new migrations (if needed):
```bash
dotnet ef migrations add MigrationName
dotnet ef database update
```

## Running the Service

The service will automatically:
1. Apply any pending migrations on startup
2. Connect to the existing database
3. Use EF Core for all database operations

No changes to the database schema are required - the EF Core models map to the existing tables.