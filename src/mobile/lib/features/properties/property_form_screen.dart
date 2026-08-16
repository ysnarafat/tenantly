import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/db/tables.dart';
import 'properties_providers.dart';

/// Create/edit form for a property. When [propertyId] is null this creates a
/// new row; otherwise it loads and edits the existing one. Purely local — no
/// network call involved either way.
class PropertyFormScreen extends ConsumerStatefulWidget {
  const PropertyFormScreen({super.key, this.propertyId});

  final int? propertyId;

  @override
  ConsumerState<PropertyFormScreen> createState() => _PropertyFormScreenState();
}

class _PropertyFormScreenState extends ConsumerState<PropertyFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _codeController = TextEditingController();
  final _addressController = TextEditingController();
  PropertyType _propertyType = PropertyType.residential;
  bool _initialized = false;
  bool _saving = false;

  bool get _isEditMode => widget.propertyId != null;

  @override
  void dispose() {
    _nameController.dispose();
    _codeController.dispose();
    _addressController.dispose();
    super.dispose();
  }

  void _populateFromExisting(PropertyRow property) {
    if (_initialized) return;
    _nameController.text = property.name;
    _codeController.text = property.code;
    _addressController.text = property.address;
    _propertyType = property.propertyType;
    _initialized = true;
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);

    final repo = ref.read(propertiesRepositoryProvider);
    if (_isEditMode) {
      final existing = await ref.read(
        propertyByIdProvider(widget.propertyId!).future,
      );
      if (existing != null) {
        await repo.update(
          existing.copyWith(
            name: _nameController.text.trim(),
            code: _codeController.text.trim(),
            propertyType: _propertyType,
            address: _addressController.text.trim(),
          ),
        );
      }
    } else {
      await repo.create(
        name: _nameController.text.trim(),
        code: _codeController.text.trim(),
        propertyType: _propertyType,
        address: _addressController.text.trim(),
      );
    }

    if (mounted) Navigator.of(context).pop();
  }

  Future<void> _delete() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete property'),
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
      await ref.read(propertiesRepositoryProvider).delete(widget.propertyId!);
      if (mounted) Navigator.of(context).pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isEditMode) {
      final propertyAsync = ref.watch(propertyByIdProvider(widget.propertyId!));
      return propertyAsync.when(
        loading: () =>
            const Scaffold(body: Center(child: CircularProgressIndicator())),
        error: (error, stackTrace) =>
            Scaffold(body: Center(child: Text('Failed to load: $error'))),
        data: (property) {
          if (property == null) {
            return const Scaffold(body: Center(child: Text('Not found')));
          }
          _populateFromExisting(property);
          return _buildForm();
        },
      );
    }
    return _buildForm();
  }

  Widget _buildForm() {
    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditMode ? 'Edit Property' : 'Add Property'),
        actions: [
          if (_isEditMode)
            IconButton(
              icon: const Icon(Icons.delete_outline),
              tooltip: 'Delete property',
              onPressed: _delete,
            ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            TextFormField(
              controller: _nameController,
              decoration: const InputDecoration(labelText: 'Property name'),
              validator: (value) =>
                  (value == null || value.trim().isEmpty) ? 'Required' : null,
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _codeController,
              decoration: const InputDecoration(labelText: 'Property code'),
              validator: (value) =>
                  (value == null || value.trim().isEmpty) ? 'Required' : null,
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<PropertyType>(
              initialValue: _propertyType,
              decoration: const InputDecoration(labelText: 'Property type'),
              items: PropertyType.values
                  .map(
                    (type) =>
                        DropdownMenuItem(value: type, child: Text(type.name)),
                  )
                  .toList(),
              onChanged: (value) {
                if (value != null) setState(() => _propertyType = value);
              },
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _addressController,
              decoration: const InputDecoration(labelText: 'Address'),
              validator: (value) =>
                  (value == null || value.trim().isEmpty) ? 'Required' : null,
              maxLines: 2,
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
