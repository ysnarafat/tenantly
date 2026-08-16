import 'package:drift/native.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tenantly_mobile/app.dart';
import 'package:tenantly_mobile/core/db/database.dart';
import 'package:tenantly_mobile/core/db/db_provider.dart';

void main() {
  testWidgets(
    'creating a property through the form makes it appear in the list',
    (tester) async {
      final db = AppDatabase.forTesting(NativeDatabase.memory());
      addTearDown(db.close);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [appDatabaseProvider.overrideWithValue(db)],
          child: const TenantlyMobileApp(),
        ),
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('No properties yet. Tap + to add one.'), findsOneWidget);

      await tester.tap(find.byType(FloatingActionButton));
      await tester.pumpAndSettle();

      expect(find.text('Add Property'), findsOneWidget);

      final fields = find.byType(TextFormField);
      await tester.enterText(fields.at(0), 'Sunset Apartments');
      await tester.enterText(fields.at(1), 'PROP-001');
      await tester.enterText(fields.at(2), '123 Main St');

      await tester.tap(find.widgetWithText(FilledButton, 'Save'));
      // Not pumpAndSettle(): the properties list screen (with its
      // never-completing live stream) stays mounted underneath the pushed
      // form the whole time, same as widget_test.dart — bounded pumps stand
      // in for it here too. The pop's default page-transition animation
      // takes ~300ms, so give it enough runway to fully finish.
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('Add Property'), findsNothing);
      expect(find.text('Sunset Apartments'), findsOneWidget);
      expect(find.textContaining('PROP-001'), findsOneWidget);
      expect(find.text('No properties yet. Tap + to add one.'), findsNothing);

      await tester.runAsync(() async {
        await tester.pumpWidget(const SizedBox());
        await Future<void>.delayed(Duration.zero);
      });
    },
  );

  testWidgets('validation blocks saving an empty form', (tester) async {
    final db = AppDatabase.forTesting(NativeDatabase.memory());
    addTearDown(db.close);

    await tester.pumpWidget(
      ProviderScope(
        overrides: [appDatabaseProvider.overrideWithValue(db)],
        child: const TenantlyMobileApp(),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    await tester.tap(find.byType(FloatingActionButton));
    await tester.pumpAndSettle();

    await tester.tap(find.widgetWithText(FilledButton, 'Save'));
    await tester.pump();

    // Still on the form (didn't navigate away) and flagging every required
    // field, since the DB was never actually written to.
    expect(find.text('Add Property'), findsOneWidget);
    expect(find.text('Required'), findsNWidgets(3));

    await tester.runAsync(() async {
      await tester.pumpWidget(const SizedBox());
      await Future<void>.delayed(Duration.zero);
    });
  });
}
