import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/db/db_provider.dart';

class LeaseWithDetails {
  LeaseWithDetails({
    required this.lease,
    required this.unitNumber,
    required this.tenantName,
  });

  final LeaseRow lease;
  final String unitNumber;
  final String tenantName;
}

/// Live query over Leases joined with their Unit's number and Tenant's name.
final leasesStreamProvider = StreamProvider<List<LeaseWithDetails>>((ref) {
  final db = ref.watch(appDatabaseProvider);
  final query = db.select(db.leases).join([
    innerJoin(db.units, db.units.id.equalsExp(db.leases.unitId)),
    innerJoin(db.tenants, db.tenants.id.equalsExp(db.leases.tenantId)),
  ])..orderBy([OrderingTerm.desc(db.leases.startDate)]);

  return query.watch().map(
    (rows) => rows
        .map(
          (row) => LeaseWithDetails(
            lease: row.readTable(db.leases),
            unitNumber: row.readTable(db.units).unitNumber,
            tenantName: row.readTable(db.tenants).name,
          ),
        )
        .toList(),
  );
});

final leaseByIdProvider = StreamProvider.family<LeaseRow?, int>((ref, id) {
  final db = ref.watch(appDatabaseProvider);
  return (db.select(
    db.leases,
  )..where((t) => t.id.equals(id))).watchSingleOrNull();
});

class LeasesRepository {
  LeasesRepository(this._db);

  final AppDatabase _db;

  Future<int> create({
    required int unitId,
    required int tenantId,
    required double monthlyRent,
    required DateTime startDate,
    DateTime? endDate,
    bool active = true,
  }) {
    return _db
        .into(_db.leases)
        .insert(
          LeasesCompanion.insert(
            unitId: unitId,
            tenantId: tenantId,
            monthlyRent: monthlyRent,
            startDate: startDate,
            endDate: Value(endDate),
            active: Value(active),
          ),
        );
  }

  Future<void> update(LeaseRow lease) {
    return _db
        .update(_db.leases)
        .replace(lease.copyWith(updatedAt: DateTime.now()));
  }

  Future<void> delete(int id) {
    return (_db.delete(_db.leases)..where((t) => t.id.equals(id))).go();
  }
}

final leasesRepositoryProvider = Provider<LeasesRepository>((ref) {
  return LeasesRepository(ref.watch(appDatabaseProvider));
});
