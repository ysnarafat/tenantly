import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/db/tables.dart';
import 'properties_providers.dart';

class PropertiesListScreen extends ConsumerWidget {
  const PropertiesListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final propertiesAsync = ref.watch(propertiesStreamProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Properties')),
      body: propertiesAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stackTrace) =>
            Center(child: Text('Failed to load properties: $error')),
        data: (properties) {
          if (properties.isEmpty) {
            return const Center(
              child: Text('No properties yet. Tap + to add one.'),
            );
          }
          return ListView.builder(
            itemCount: properties.length,
            itemBuilder: (context, index) {
              final property = properties[index];
              return Dismissible(
                key: ValueKey(property.id),
                direction: DismissDirection.endToStart,
                background: Container(
                  color: Theme.of(context).colorScheme.errorContainer,
                  alignment: Alignment.centerRight,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: const Icon(Icons.delete_outline),
                ),
                onDismissed: (_) {
                  ref.read(propertiesRepositoryProvider).delete(property.id);
                },
                child: ListTile(
                  title: Text(property.name),
                  subtitle: Text('${property.code} · ${property.address}'),
                  trailing: Text(_propertyTypeLabel(property.propertyType)),
                  onTap: () => context.push('/properties/${property.id}/edit'),
                ),
              );
            },
          );
        },
      ),
      floatingActionButton: FloatingActionButton(
        heroTag: 'properties-fab',
        onPressed: () => context.push('/properties/new'),
        tooltip: 'Add property',
        child: const Icon(Icons.add),
      ),
    );
  }

  String _propertyTypeLabel(PropertyType type) {
    switch (type) {
      case PropertyType.residential:
        return 'Residential';
      case PropertyType.commercial:
        return 'Commercial';
      case PropertyType.mixed:
        return 'Mixed';
    }
  }
}
