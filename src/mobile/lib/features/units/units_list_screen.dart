import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/db/tables.dart';
import 'units_providers.dart';

class UnitsListScreen extends ConsumerWidget {
  const UnitsListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final unitsAsync = ref.watch(unitsStreamProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Units')),
      body: unitsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stackTrace) =>
            Center(child: Text('Failed to load units: $error')),
        data: (units) {
          if (units.isEmpty) {
            return const Center(child: Text('No units yet. Tap + to add one.'));
          }
          return ListView.builder(
            itemCount: units.length,
            itemBuilder: (context, index) {
              final entry = units[index];
              final unit = entry.unit;
              return Dismissible(
                key: ValueKey(unit.id),
                direction: DismissDirection.endToStart,
                background: Container(
                  color: Theme.of(context).colorScheme.errorContainer,
                  alignment: Alignment.centerRight,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: const Icon(Icons.delete_outline),
                ),
                onDismissed: (_) {
                  ref.read(unitsRepositoryProvider).delete(unit.id);
                },
                child: ListTile(
                  title: Text(unit.unitNumber),
                  subtitle: Text(
                    unit.floor != null
                        ? '${entry.buildingName} · Floor ${unit.floor}'
                        : entry.buildingName,
                  ),
                  trailing: Text(_unitTypeLabel(unit.unitType)),
                  onTap: () => context.push('/units/${unit.id}/edit'),
                ),
              );
            },
          );
        },
      ),
      floatingActionButton: FloatingActionButton(
        heroTag: 'units-fab',
        onPressed: () => context.push('/units/new'),
        tooltip: 'Add unit',
        child: const Icon(Icons.add),
      ),
    );
  }

  String _unitTypeLabel(UnitType type) {
    switch (type) {
      case UnitType.shop:
        return 'Shop';
      case UnitType.apartment:
        return 'Apartment';
      case UnitType.office:
        return 'Office';
      case UnitType.parking:
        return 'Parking';
      case UnitType.storage:
        return 'Storage';
      case UnitType.other:
        return 'Other';
    }
  }
}
