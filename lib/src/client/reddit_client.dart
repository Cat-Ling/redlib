import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import '../auth/session_manager.dart';

class RedditClient {
  final SessionManager _sessionManager;
  static const String _apiBaseUrl = 'https://oauth.reddit.com';

  RedditClient(this._sessionManager);

  Future<dynamic> _get(String path) async {
    final session = await _sessionManager.getSession();
    final url = Uri.parse('$_apiBaseUrl$path');

    final response = await http.get(url, headers: session.headers);

    if (response.statusCode == 200) {
      return json.decode(response.body);
    } else {
      throw Exception('Failed to load data from $path: ${response.statusCode} ${response.body}');
    }
  }

  Future<dynamic> getFrontPage({String sort = 'hot'}) async {
    return _get('/$sort.json');
  }

  Future<dynamic> getSubreddit(String subreddit, {String sort = 'hot'}) async {
    return _get('/r/$subreddit/$sort.json');
  }

  Future<dynamic> getComments(String subreddit, String postId) async {
    return _get('/r/$subreddit/comments/$postId.json');
  }
}
