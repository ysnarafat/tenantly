import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/db/db_provider.dart';
import '../../core/db/tables.dart';

class UnitWithBuilding {
  UnitWithBuilding({required this.unit, required this.buildingName});

  final UnitRow unit;
  final String buildingName;
}

/// Live query over Units joined with their parent Building's name.
final unitsStreamProvider = StreamProvider<List<UnitWithBuilding>>((ref) {
  final db = ref.watch(appDatabaseProvider);
  final query = db.select(db.units).join([
    innerJoin(db.buildings, db.buildings.id.equalsExp(db.units.buildingId)),
  ])..orderBy([OrderingTerm.asc(db.units.unitNumber)]);

  return query.watch().map(
    (rows) => rows
        .map(
          (row) => UnitWithBuilding(
            unit: row.readTable(db.units),
            buildingName: row.readTable(db.buildings).name,
          ),
        )
        .toList(),
  );
});

final unitByIdProvider = StreamProvider.family<UnitRow?, int>((ref, id) {
  final db = ref.watch(appDatabaseProvider);
  return (db.select(
    db.units,
  )..where((t) => t.id.equals(id))).watchSingleOrNull();
});

class UnitsRepository {
  UnitsRepository(this._db);

  final AppDatabase _db;

  Future<int> create({
    required int buildingId,
    required String unitNumber,
    required UnitType unitType,
    int? floor,
  }) {
    return _db
        .into(_db.units)
        .insert(
          UnitsCompanion.insert(
            buildingId: buildingId,
            unitNumber: unitNumber,
            unitType: unitType,
            floor: Value(floor),
          ),
        );
  }

  Future<void> update(UnitRow unit) {
    return _db
        .update(_db.units)
        .replace(unit.copyWith(updatedAt: DateTime.now()));
  }

  Future<void> delete(int id) {
    return (_db.delete(_db.units)..where((t) => t.id.equals(id))).go();
  }
}

final unitsRepositoryProvider = Provider<UnitsRepository>((ref) {
  return UnitsRepository(ref.watch(appDatabaseProvider));
});
