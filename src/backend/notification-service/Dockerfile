# Build stage
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

# Copy project file first for better caching
COPY *.csproj .

# Restore dependencies (this layer will be cached if .csproj doesn't change)
RUN dotnet restore TenantlyNotificationService.csproj --no-cache

# Copy source code
COPY . .

# Build and publish in one step for better compatibility. Target the
# deployable project explicitly rather than letting `dotnet publish`
# auto-discover the .sln, which also pulls in TenantlyNotificationService.Tests
# — a project never restored in this image and not meant to be shipped.
RUN dotnet publish TenantlyNotificationService.csproj -c Release -o /app/publish \
    --no-restore \
    --verbosity minimal

# Runtime stage
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS runtime
WORKDIR /app

# Create a dedicated non-root user rather than assuming one is built into this
# base image tag.
RUN groupadd -r appgroup && useradd -r -u 1001 -g appgroup appuser

# Create logs directory
RUN mkdir -p /app/logs

# Copy published application
COPY --from=build /app/publish .

RUN chown -R appuser:appgroup /app

USER appuser

ENTRYPOINT ["dotnet", "TenantlyNotificationService.dll"]