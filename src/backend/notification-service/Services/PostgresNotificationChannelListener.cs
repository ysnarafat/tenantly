using Microsoft.Extensions.Configuration;
using Npgsql;
using TenantlyNotificationService.Configuration;

namespace TenantlyNotificationService.Services;

/// <summary>
/// Listens on the Postgres "notification_queue_channel" (populated by a DB
/// trigger on notification_queue inserts) using LISTEN/NOTIFY, so the worker
/// can react to new rows immediately instead of waiting for its next poll.
/// Holds a single dedicated connection outside of EF Core's pool, since
/// LISTEN is scoped to the connection that issued it.
/// </summary>
public class PostgresNotificationChannelListener : INotificationChannelListener, IAsyncDisposable
{
    private const string ChannelName = "notification_queue_channel";

    private readonly string _connectionString;
    private readonly ILogger<PostgresNotificationChannelListener> _logger;
    private NpgsqlConnection? _connection;

    public PostgresNotificationChannelListener(IConfiguration configuration, ILogger<PostgresNotificationChannelListener> logger)
    {
        _connectionString = ConnectionStringResolver.Resolve(configuration);
        _logger = logger;
    }

    public async Task<bool> WaitForNotificationAsync(TimeSpan timeout, CancellationToken cancellationToken)
    {
        NpgsqlConnection connection;
        try
        {
            connection = await EnsureConnectionAsync(cancellationToken);
        }
        catch (Exception ex) when (!cancellationToken.IsCancellationRequested)
        {
            _logger.LogWarning(ex, "Could not establish LISTEN connection; falling back to polling only for now");
            await DisposeConnectionAsync();
            await Task.Delay(timeout, cancellationToken);
            return false;
        }

        using var timeoutCts = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
        timeoutCts.CancelAfter(timeout);

        try
        {
            await connection.WaitAsync(timeoutCts.Token);
            return true;
        }
        catch (OperationCanceledException) when (!cancellationToken.IsCancellationRequested)
        {
            return false;
        }
        catch (Exception ex) when (!cancellationToken.IsCancellationRequested)
        {
            _logger.LogWarning(ex, "LISTEN connection dropped; will reconnect on the next cycle");
            await DisposeConnectionAsync();
            return false;
        }
    }

    private async Task<NpgsqlConnection> EnsureConnectionAsync(CancellationToken cancellationToken)
    {
        if (_connection is { State: System.Data.ConnectionState.Open })
        {
            return _connection;
        }

        await DisposeConnectionAsync();

        var connection = new NpgsqlConnection(_connectionString);
        await connection.OpenAsync(cancellationToken);

        await using (var command = new NpgsqlCommand($"LISTEN {ChannelName};", connection))
        {
            await command.ExecuteNonQueryAsync(cancellationToken);
        }

        _logger.LogInformation("Listening for Postgres notifications on channel {Channel}", ChannelName);
        _connection = connection;
        return connection;
    }

    private async Task DisposeConnectionAsync()
    {
        if (_connection != null)
        {
            await _connection.DisposeAsync();
            _connection = null;
        }
    }

    public async ValueTask DisposeAsync() => await DisposeConnectionAsync();
}
