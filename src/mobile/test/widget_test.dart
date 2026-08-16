import 'package:drift/native.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tenantly_mobile/app.dart';
import 'package:tenantly_mobile/core/db/database.dart';
import 'package:tenantly_mobile/core/db/db_provider.dart';

void main() {
  testWidgets('app opens straight to the Properties tab, no auth gate', (
    WidgetTester tester,
  ) async {
    // Override with an in-memory DB — the real provider opens a file via
    // path_provider, whose platform channel isn't available in widget tests.
    final testDb = AppDatabase.forTesting(NativeDatabase.memory());
    addTearDown(testDb.close);

    await tester.pumpWidget(
      ProviderScope(
        overrides: [appDatabaseProvider.overrideWithValue(testDb)],
        child: const TenantlyMobileApp(),
      ),
    );
    // Not pumpAndSettle(): the properties list watches a live Drift stream
    // that never completes, so "settled" (no more frames scheduled) never
    // arrives — a couple of bounded pumps is enough for the first snapshot.
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(find.text('Properties'), findsWidgets);
    expect(find.text('No properties yet. Tap + to add one.'), findsOneWidget);

    // Drift's stream cancellation schedules a zero-duration Timer when the
    // widget tree (and with it, ProviderScope's StreamProvider) is disposed.
    // The test binding runs pump()/pumpWidget() inside a FakeAsync zone,
    // where that timer never actually fires no matter how many times it's
    // pumped afterward — runAsync() steps outside FakeAsync into a real
    // async zone so the pending timer can actually run to completion.
    await tester.runAsync(() async {
      await tester.pumpWidget(const SizedBox());
      await Future<void>.delayed(Duration.zero);
    });
  });
}
