import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/db/tables.dart';
import '../properties/properties_providers.dart';
import 'buildings_providers.dart';

/// Create/edit form for a building. Mirrors [PropertyFormScreen]'s shape —
/// the only addition is the required Property picker, since a building
/// can't exist without a parent property.
class BuildingFormScreen extends ConsumerStatefulWidget {
  const BuildingFormScreen({super.key, this.buildingId});

  final int? buildingId;

  @override
  ConsumerState<BuildingFormScreen> createState() => _BuildingFormScreenState();
}

class _BuildingFormScreenState extends ConsumerState<BuildingFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _codeController = TextEditingController();
  BuildingType _buildingType = BuildingType.residential;
  int? _propertyId;
  bool _initialized = false;
  bool _saving = false;

  bool get _isEditMode => widget.buildingId != null;

  @override
  void dispose() {
    _nameController.dispose();
    _codeController.dispose();
    super.dispose();
  }

  void _populateFromExisting(BuildingRow building) {
    if (_initialized) return;
    _nameController.text = building.name;
    _codeController.text = building.code;
    _buildingType = building.buildingType;
    _propertyId = building.propertyId;
    _initialized = true;
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    if (_propertyId == null) return;
    setState(() => _saving = true);

    final repo = ref.read(buildingsRepositoryProvider);
    if (_isEditMode) {
      final existing = await ref.read(
        buildingByIdProvider(widget.buildingId!).future,
      );
      if (existing != null) {
        await repo.update(
          existing.copyWith(
            propertyId: _propertyId!,
            name: _nameController.text.trim(),
            code: _codeController.text.trim(),
            buildingType: _buildingType,
          ),
        );
      }
    } else {
      await repo.create(
        propertyId: _propertyId!,
        name: _nameController.text.trim(),
        code: _codeController.text.trim(),
        buildingType: _buildingType,
      );
    }

    if (mounted) Navigator.of(context).pop();
  }

  Future<void> _delete() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete building'),
        content: const Text(
          'This also deletes every unit under this building. This cannot be undone.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await ref.read(buildingsRepositoryProvider).delete(widget.buildingId!);
      if (mounted) Navigator.of(context).pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isEditMode) {
      final buildingAsync = ref.watch(buildingByIdProvider(widget.buildingId!));
      return buildingAsync.when(
        loading: () =>
            const Scaffold(body: Center(child: CircularProgressIndicator())),
        error: (error, stackTrace) =>
            Scaffold(body: Center(child: Text('Failed to load: $error'))),
        data: (building) {
          if (building == null) {
            return const Scaffold(body: Center(child: Text('Not found')));
          }
          _populateFromExisting(building);
          return _buildForm();
        },
      );
    }
    return _buildForm();
  }

  Widget _buildForm() {
    final propertiesAsync = ref.watch(propertiesStreamProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditMode ? 'Edit Building' : 'Add Building'),
        actions: [
          if (_isEditMode)
            IconButton(
              icon: const Icon(Icons.delete_outline),
              tooltip: 'Delete building',
              onPressed: _delete,
            ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            propertiesAsync.when(
              loading: () => const LinearProgressIndicator(),
              error: (error, stackTrace) => Text('Failed to load properties'),
              data: (properties) {
                if (properties.isEmpty) {
                  return const Text(
                    'Add a property first before creating a building.',
                  );
                }
                _propertyId ??= properties.first.id;
                return DropdownButtonFormField<int>(
                  initialValue: _propertyId,
                  decoration: const InputDecoration(labelText: 'Property'),
                  items: properties
                      .map(
                        (p) =>
                            DropdownMenuItem(value: p.id, child: Text(p.name)),
                      )
                      .toList(),
                  onChanged: (value) => setState(() => _propertyId = value),
                  validator: (value) => value == null ? 'Required' : null,
                );
              },
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _nameController,
              decoration: const InputDecoration(labelText: 'Building name'),
              validator: (value) =>
                  (value == null || value.trim().isEmpty) ? 'Required' : null,
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _codeController,
              decoration: const InputDecoration(labelText: 'Building code'),
              validator: (value) =>
                  (value == null || value.trim().isEmpty) ? 'Required' : null,
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<BuildingType>(
              initialValue: _buildingType,
              decoration: const InputDecoration(labelText: 'Building type'),
              items: BuildingType.values
                  .map(
                    (type) =>
                        DropdownMenuItem(value: type, child: Text(type.name)),
                  )
                  .toList(),
              onChanged: (value) {
                if (value != null) setState(() => _buildingType = value);
              },
            ),
            const SizedBox(height: 24),
            FilledButton(
              onPressed: _saving ? null : _save,
              child: Text(_saving ? 'Saving...' : 'Save'),
            ),
          ],
        ),
      ),
    );
  }
}
