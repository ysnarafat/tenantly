using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging.Abstractions;
using Npgsql;
using TenantlyNotificationService.Services;

namespace TenantlyNotificationService.Tests;

/// <summary>
/// Exercises the real Postgres LISTEN/NOTIFY plumbing end to end: a trigger
/// (the same SQL shipped in migration 000017_notification_queue_notify) fires
/// pg_notify on insert, and PostgresNotificationChannelListener must wake up
/// from it. Requires a real, reachable Postgres instance - unlike the rest of
/// the suite, which runs against Sqlite/mocks with no external dependency.
///
/// Reuses the same Postgres instance/credentials as the Go integration tests
/// documented in CLAUDE.md (localhost:5432, user "user", password "p@ssw0rd",
/// database "tenantly_test") so a dev who already has that running for
/// `go test ./internal/repositories/...` needs no extra setup. Override with
/// NOTIFICATION_SERVICE_TEST_CONNECTION_STRING if your test DB differs.
///
/// Run with: dotnet test --filter Category=Integration
/// Excluded from the fast unit-test run with: dotnet test --filter Category!=Integration
/// </summary>
[Trait("Category", "Integration")]
public class PostgresNotificationChannelListenerIntegrationTests : IAsyncLifetime
{
    private static readonly string ConnectionString =
        Environment.GetEnvironmentVariable("NOTIFICATION_SERVICE_TEST_CONNECTION_STRING")
        ?? "Host=localhost;Port=5432;Database=tenantly_test;Username=user;Password=p@ssw0rd";

    private const string TestTableName = "listener_test_notification_queue";
    private const string TriggerFunctionName = "notify_notification_queue_insert";

    private NpgsqlConnection _adminConnection = null!;

    public async Task InitializeAsync()
    {
        _adminConnection = new NpgsqlConnection(ConnectionString);
        await _adminConnection.OpenAsync();

        // A minimal stand-in for notification_queue: the trigger only ever
        // touches NEW.id, so we don't need the real table's tenant/unit FKs.
        await ExecuteAsync($"""
            DROP TABLE IF EXISTS {TestTableName};
            CREATE TABLE {TestTableName} (id SERIAL PRIMARY KEY);

            CREATE OR REPLACE FUNCTION {TriggerFunctionName}()
            RETURNS TRIGGER AS $$
            BEGIN
                PERFORM pg_notify('notification_queue_channel', NEW.id::text);
                RETURN NEW;
            END;
            $$ LANGUAGE plpgsql;

            DROP TRIGGER IF EXISTS notify_{TestTableName}_trigger ON {TestTableName};
            CREATE TRIGGER notify_{TestTableName}_trigger
                AFTER INSERT ON {TestTableName}
                FOR EACH ROW EXECUTE FUNCTION {TriggerFunctionName}();
            """);
    }

    public async Task DisposeAsync()
    {
        await ExecuteAsync($"DROP TABLE IF EXISTS {TestTableName};");
        await _adminConnection.DisposeAsync();
    }

    private async Task ExecuteAsync(string sql)
    {
        await using var command = new NpgsqlCommand(sql, _adminConnection);
        await command.ExecuteNonQueryAsync();
    }

    private static PostgresNotificationChannelListener CreateListener()
    {
        var configuration = new ConfigurationBuilder()
            .AddInMemoryCollection(new Dictionary<string, string?>
            {
                ["ConnectionStrings:DefaultConnection"] = ConnectionString,
            })
            .Build();

        return new PostgresNotificationChannelListener(configuration, NullLogger<PostgresNotificationChannelListener>.Instance);
    }

    [Fact]
    public async Task WaitForNotificationAsync_ReturnsTrue_WhenTriggerFiresOnInsert()
    {
        await using var listener = CreateListener();

        // Opens the dedicated connection and issues LISTEN. Awaiting this
        // (even though nothing has been NOTIFYed yet) is a synchronization
        // point: it guarantees the subscription is registered with Postgres
        // before we insert, so there's no race between "start listening" and
        // "trigger fires" below.
        await listener.WaitForNotificationAsync(TimeSpan.FromMilliseconds(50), CancellationToken.None);

        var waitTask = listener.WaitForNotificationAsync(TimeSpan.FromSeconds(10), CancellationToken.None);

        await ExecuteAsync($"INSERT INTO {TestTableName} DEFAULT VALUES;");

        var notified = await waitTask;

        Assert.True(notified);
    }

    [Fact]
    public async Task WaitForNotificationAsync_ReturnsFalse_WhenNoNotificationArrivesBeforeTimeout()
    {
        await using var listener = CreateListener();

        var notified = await listener.WaitForNotificationAsync(TimeSpan.FromMilliseconds(300), CancellationToken.None);

        Assert.False(notified);
    }
}
