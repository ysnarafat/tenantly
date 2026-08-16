import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/db/db_provider.dart';
import '../../core/db/tables.dart';

class BuildingWithProperty {
  BuildingWithProperty({required this.building, required this.propertyName});

  final BuildingRow building;
  final String propertyName;
}

/// Live query over Buildings joined with their parent Property's name — the
/// list screen wants "PropertyName · BuildingName", not just a bare
/// building record, same as the web app's property-detail buildings list.
final buildingsStreamProvider = StreamProvider<List<BuildingWithProperty>>((
  ref,
) {
  final db = ref.watch(appDatabaseProvider);
  final query = db.select(db.buildings).join([
    innerJoin(
      db.properties,
      db.properties.id.equalsExp(db.buildings.propertyId),
    ),
  ])..orderBy([OrderingTerm.asc(db.buildings.name)]);

  return query.watch().map(
    (rows) => rows
        .map(
          (row) => BuildingWithProperty(
            building: row.readTable(db.buildings),
            propertyName: row.readTable(db.properties).name,
          ),
        )
        .toList(),
  );
});

final buildingByIdProvider = StreamProvider.family<BuildingRow?, int>((
  ref,
  id,
) {
  final db = ref.watch(appDatabaseProvider);
  return (db.select(
    db.buildings,
  )..where((t) => t.id.equals(id))).watchSingleOrNull();
});

class BuildingsRepository {
  BuildingsRepository(this._db);

  final AppDatabase _db;

  Future<int> create({
    required int propertyId,
    required String name,
    required String code,
    required BuildingType buildingType,
  }) {
    return _db
        .into(_db.buildings)
        .insert(
          BuildingsCompanion.insert(
            propertyId: propertyId,
            name: name,
            code: code,
            buildingType: buildingType,
          ),
        );
  }

  Future<void> update(BuildingRow building) {
    return _db
        .update(_db.buildings)
        .replace(building.copyWith(updatedAt: DateTime.now()));
  }

  Future<void> delete(int id) {
    return (_db.delete(_db.buildings)..where((t) => t.id.equals(id))).go();
  }
}

final buildingsRepositoryProvider = Provider<BuildingsRepository>((ref) {
  return BuildingsRepository(ref.watch(appDatabaseProvider));
});
