# Build stage
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

# Copy project file first for better caching
COPY *.csproj .

# Restore dependencies (this layer will be cached if .csproj doesn't change)
RUN dotnet restore --no-cache

# Copy source code
COPY . .

# Build and publish in one step for better compatibility
RUN dotnet publish -c Release -o /app/publish \
    --no-restore \
    --verbosity minimal

# Runtime stage
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS runtime
WORKDIR /app

# Create logs directory
RUN mkdir -p /app/logs

# Copy published application
COPY --from=build /app/publish .

ENTRYPOINT ["dotnet", "TenantlyNotificationService.dll"]