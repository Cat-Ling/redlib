import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'src/auth/session_manager.dart';
import 'src/client/reddit_client.dart';
import 'src/screens/home_screen.dart';

void main() {
  runApp(MyApp());
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Provider<RedditClient>(
      create: (_) => RedditClient(SessionManager()),
      child: MaterialApp(
        title: 'Reddit Client',
        theme: ThemeData(
          primarySwatch: Colors.blue,
        ),
        home: HomeScreen(),
      ),
    );
  }
}
