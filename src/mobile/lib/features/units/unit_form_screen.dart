import 'package:drift/drift.dart' show Value;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/db/tables.dart';
import '../buildings/buildings_providers.dart';
import 'units_providers.dart';

class UnitFormScreen extends ConsumerStatefulWidget {
  const UnitFormScreen({super.key, this.unitId});

  final int? unitId;

  @override
  ConsumerState<UnitFormScreen> createState() => _UnitFormScreenState();
}

class _UnitFormScreenState extends ConsumerState<UnitFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _unitNumberController = TextEditingController();
  final _floorController = TextEditingController();
  UnitType _unitType = UnitType.apartment;
  int? _buildingId;
  bool _initialized = false;
  bool _saving = false;

  bool get _isEditMode => widget.unitId != null;

  @override
  void dispose() {
    _unitNumberController.dispose();
    _floorController.dispose();
    super.dispose();
  }

  void _populateFromExisting(UnitRow unit) {
    if (_initialized) return;
    _unitNumberController.text = unit.unitNumber;
    _floorController.text = unit.floor?.toString() ?? '';
    _unitType = unit.unitType;
    _buildingId = unit.buildingId;
    _initialized = true;
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    if (_buildingId == null) return;
    setState(() => _saving = true);

    final floor = _floorController.text.trim().isEmpty
        ? null
        : int.tryParse(_floorController.text.trim());
    final repo = ref.read(unitsRepositoryProvider);
    if (_isEditMode) {
      final existing = await ref.read(unitByIdProvider(widget.unitId!).future);
      if (existing != null) {
        await repo.update(
          existing.copyWith(
            buildingId: _buildingId!,
            unitNumber: _unitNumberController.text.trim(),
            unitType: _unitType,
            floor: Value(floor),
          ),
        );
      }
    } else {
      await repo.create(
        buildingId: _buildingId!,
        unitNumber: _unitNumberController.text.trim(),
        unitType: _unitType,
        floor: floor,
      );
    }

    if (mounted) Navigator.of(context).pop();
  }

  Future<void> _delete() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete unit'),
        content: const Text('This cannot be undone.'),
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
      await ref.read(unitsRepositoryProvider).delete(widget.unitId!);
      if (mounted) Navigator.of(context).pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isEditMode) {
      final unitAsync = ref.watch(unitByIdProvider(widget.unitId!));
      return unitAsync.when(
        loading: () =>
            const Scaffold(body: Center(child: CircularProgressIndicator())),
        error: (error, stackTrace) =>
            Scaffold(body: Center(child: Text('Failed to load: $error'))),
        data: (unit) {
          if (unit == null) {
            return const Scaffold(body: Center(child: Text('Not found')));
          }
          _populateFromExisting(unit);
          return _buildForm();
        },
      );
    }
    return _buildForm();
  }

  Widget _buildForm() {
    final buildingsAsync = ref.watch(buildingsStreamProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditMode ? 'Edit Unit' : 'Add Unit'),
        actions: [
          if (_isEditMode)
            IconButton(
              icon: const Icon(Icons.delete_outline),
              tooltip: 'Delete unit',
              onPressed: _delete,
            ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            buildingsAsync.when(
              loading: () => const LinearProgressIndicator(),
              error: (error, stackTrace) =>
                  const Text('Failed to load buildings'),
              data: (buildings) {
                if (buildings.isEmpty) {
                  return const Text(
                    'Add a building first before creating a unit.',
                  );
                }
                _buildingId ??= buildings.first.building.id;
                return DropdownButtonFormField<int>(
                  initialValue: _buildingId,
                  decoration: const InputDecoration(labelText: 'Building'),
                  items: buildings
                      .map(
                        (b) => DropdownMenuItem(
                          value: b.building.id,
                          child: Text('${b.propertyName} · ${b.building.name}'),
                        ),
                      )
                      .toList(),
                  onChanged: (value) => setState(() => _buildingId = value),
                  validator: (value) => value == null ? 'Required' : null,
                );
              },
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _unitNumberController,
              decoration: const InputDecoration(labelText: 'Unit number'),
              validator: (value) =>
                  (value == null || value.trim().isEmpty) ? 'Required' : null,
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<UnitType>(
              initialValue: _unitType,
              decoration: const InputDecoration(labelText: 'Unit type'),
              items: UnitType.values
                  .map(
                    (type) =>
                        DropdownMenuItem(value: type, child: Text(type.name)),
                  )
                  .toList(),
              onChanged: (value) {
                if (value != null) setState(() => _unitType = value);
              },
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _floorController,
              decoration: const InputDecoration(labelText: 'Floor (optional)'),
              keyboardType: TextInputType.number,
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
