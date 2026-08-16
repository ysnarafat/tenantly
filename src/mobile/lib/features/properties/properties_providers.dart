import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/db/db_provider.dart';
import '../../core/db/tables.dart';

/// Live query over the local Properties table — updates automatically on
/// every insert/update/delete, with no network round-trip involved.
final propertiesStreamProvider = StreamProvider<List<PropertyRow>>((ref) {
  final db = ref.watch(appDatabaseProvider);
  return db.select(db.properties).watch();
});

final propertyByIdProvider = StreamProvider.family<PropertyRow?, int>((
  ref,
  id,
) {
  final db = ref.watch(appDatabaseProvider);
  return (db.select(
    db.properties,
  )..where((t) => t.id.equals(id))).watchSingleOrNull();
});

class PropertiesRepository {
  PropertiesRepository(this._db);

  final AppDatabase _db;

  Future<int> create({
    required String name,
    required String code,
    required PropertyType propertyType,
    required String address,
  }) {
    return _db
        .into(_db.properties)
        .insert(
          PropertiesCompanion.insert(
            name: name,
            code: code,
            propertyType: propertyType,
            address: address,
          ),
        );
  }

  Future<void> update(PropertyRow property) {
    return _db
        .update(_db.properties)
        .replace(property.copyWith(updatedAt: DateTime.now()));
  }

  Future<void> delete(int id) {
    return (_db.delete(_db.properties)..where((t) => t.id.equals(id))).go();
  }
}

final propertiesRepositoryProvider = Provider<PropertiesRepository>((ref) {
  return PropertiesRepository(ref.watch(appDatabaseProvider));
});
