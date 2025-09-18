import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mockito/annotations.dart';
import 'package:mockito/mockito.dart';
import 'package:provider/provider.dart';
import 'package:reddit_client/src/client/reddit_client.dart';
import 'package:reddit_client/src/models/post.dart';
import 'package:reddit_client/main.dart';
import 'package:reddit_client/src/auth/oauth.dart';
import 'package:reddit_client/src/auth/session_manager.dart';

import 'widget_test.mocks.dart';

@GenerateMocks([RedditClient, SessionManager])
void main() {
  testWidgets('HomeScreen displays posts', (WidgetTester tester) async {
    final mockRedditClient = MockRedditClient();
    final mockSessionManager = MockSessionManager();

    // Stub the session manager to avoid actual auth calls
    when(mockSessionManager.getSession()).thenAnswer((_) async => OauthSession(accessToken: 'fake_token', expiresIn: 3600, headers: {}));

    final posts = [
      Post(id: '1', title: 'Post 1', author: 'author1', subreddit: 'sub1', score: 10, selftext: '', url: '', numComments: 0),
      Post(id: '2', title: 'Post 2', author: 'author2', subreddit: 'sub2', score: 20, selftext: '', url: '', numComments: 0),
    ];

    when(mockRedditClient.getFrontPage()).thenAnswer((_) async => {
      'data': {
        'children': posts.map((p) => {'data': {'id': p.id, 'title': p.title, 'author': p.author, 'subreddit': p.subreddit, 'score': p.score, 'selftext': p.selftext, 'url': p.url, 'num_comments': p.numComments}}).toList()
      }
    });

    await tester.pumpWidget(
      Provider<RedditClient>.value(
        value: mockRedditClient,
        child: MyApp(),
      ),
    );

    // Let the FutureBuilder resolve.
    await tester.pumpAndSettle();

    expect(find.text('Post 1'), findsOneWidget);
    expect(find.text('Post 2'), findsOneWidget);
  });
}
