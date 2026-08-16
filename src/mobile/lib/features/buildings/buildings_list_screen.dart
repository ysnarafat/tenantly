import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/db/tables.dart';
import 'buildings_providers.dart';

class BuildingsListScreen extends ConsumerWidget {
  const BuildingsListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final buildingsAsync = ref.watch(buildingsStreamProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Buildings')),
      body: buildingsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stackTrace) =>
            Center(child: Text('Failed to load buildings: $error')),
        data: (buildings) {
          if (buildings.isEmpty) {
            return const Center(
              child: Text('No buildings yet. Tap + to add one.'),
            );
          }
          return ListView.builder(
            itemCount: buildings.length,
            itemBuilder: (context, index) {
              final entry = buildings[index];
              final building = entry.building;
              return Dismissible(
                key: ValueKey(building.id),
                direction: DismissDirection.endToStart,
                background: Container(
                  color: Theme.of(context).colorScheme.errorContainer,
                  alignment: Alignment.centerRight,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: const Icon(Icons.delete_outline),
                ),
                onDismissed: (_) {
                  ref.read(buildingsRepositoryProvider).delete(building.id);
                },
                child: ListTile(
                  title: Text(building.name),
                  subtitle: Text('${entry.propertyName} · ${building.code}'),
                  trailing: Text(_buildingTypeLabel(building.buildingType)),
                  onTap: () => context.push('/buildings/${building.id}/edit'),
                ),
              );
            },
          );
        },
      ),
      floatingActionButton: FloatingActionButton(
        heroTag: 'buildings-fab',
        onPressed: () => context.push('/buildings/new'),
        tooltip: 'Add building',
        child: const Icon(Icons.add),
      ),
    );
  }

  String _buildingTypeLabel(BuildingType type) {
    switch (type) {
      case BuildingType.residential:
        return 'Residential';
      case BuildingType.commercial:
        return 'Commercial';
      case BuildingType.mixed:
        return 'Mixed';
    }
  }
}
