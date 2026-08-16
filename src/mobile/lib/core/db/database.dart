import 'dart:io';

import 'package:drift/drift.dart';
import 'package:drift/native.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';

import 'tables.dart';

part 'database.g.dart';

/// Local-first: this database is the single source of truth for the app.
/// There is no network/sync layer yet — every read and write goes straight
/// here (see the mobile setup plan for what's explicitly deferred).
@DriftDatabase(
  tables: [Properties, Buildings, Units, Tenants, Leases, Payments],
)
class AppDatabase extends _$AppDatabase {
  AppDatabase() : super(_openConnection());

  AppDatabase.forTesting(super.executor);

  @override
  int get schemaVersion => 1;

  // Default drift storage truncates DateTime columns to whole seconds
  // (unix epoch integer), so a create followed by an update within the same
  // second produces identical updatedAt values. Storing as ISO-8601 text
  // keeps full precision.
  @override
  DriftDatabaseOptions get options =>
      const DriftDatabaseOptions(storeDateTimeAsText: true);

  // SQLite has foreign keys OFF by default on every connection — without
  // this, the onDelete: KeyAction.cascade declared on Buildings/Units/Leases
  // silently does nothing.
  @override
  MigrationStrategy get migration => MigrationStrategy(
    beforeOpen: (details) async {
      await customStatement('PRAGMA foreign_keys = ON');
    },
  );
}

LazyDatabase _openConnection() {
  return LazyDatabase(() async {
    final dbFolder = await getApplicationDocumentsDirectory();
    final file = File(p.join(dbFolder.path, 'tenantly_mobile.sqlite'));
    return NativeDatabase.createInBackground(file);
  });
}
