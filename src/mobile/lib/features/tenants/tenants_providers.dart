import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/db/db_provider.dart';

final tenantsStreamProvider = StreamProvider<List<TenantRow>>((ref) {
  final db = ref.watch(appDatabaseProvider);
  return (db.select(
    db.tenants,
  )..orderBy([(t) => OrderingTerm.asc(t.name)])).watch();
});

final tenantByIdProvider = StreamProvider.family<TenantRow?, int>((ref, id) {
  final db = ref.watch(appDatabaseProvider);
  return (db.select(
    db.tenants,
  )..where((t) => t.id.equals(id))).watchSingleOrNull();
});

class TenantsRepository {
  TenantsRepository(this._db);

  final AppDatabase _db;

  Future<int> create({required String name, String? phone, String? nid}) {
    return _db
        .into(_db.tenants)
        .insert(
          TenantsCompanion.insert(
            name: name,
            phone: Value(phone),
            nidNumber: Value(nid),
          ),
        );
  }

  Future<void> update(TenantRow tenant) {
    return _db
        .update(_db.tenants)
        .replace(tenant.copyWith(updatedAt: DateTime.now()));
  }

  Future<void> delete(int id) {
    return (_db.delete(_db.tenants)..where((t) => t.id.equals(id))).go();
  }
}

final tenantsRepositoryProvider = Provider<TenantsRepository>((ref) {
  return TenantsRepository(ref.watch(appDatabaseProvider));
});
