import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/device.dart';

class OauthSession {
  final String accessToken;
  final int expiresIn;
  final Map<String, String> headers;

  OauthSession({
    required this.accessToken,
    required this.expiresIn,
    required this.headers,
  });
}

class Oauth {
  static const String _authEndpoint = 'https://www.reddit.com/auth/v2/oauth/access-token/loid';

  static Future<OauthSession> login() async {
    final device = Device.android();
    final authString = 'Basic ${base64.encode(utf8.encode('${device.oauthId}:'))}';

    final headers = {
      ...device.initialHeaders,
      'Authorization': authString,
    };

    final body = json.encode({
      'scopes': ['*', 'email', 'pii']
    });

    final response = await http.post(
      Uri.parse(_authEndpoint),
      headers: headers,
      body: body,
    );

    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      final sessionHeaders = <String, String>{
        'Authorization': 'Bearer ${data['access_token']}',
      };
      if (response.headers['x-reddit-loid'] != null) {
        sessionHeaders['x-reddit-loid'] = response.headers['x-reddit-loid']!;
      }
      if (response.headers['x-reddit-session'] != null) {
        sessionHeaders['x-reddit-session'] = response.headers['x-reddit-session']!;
      }

      return OauthSession(
        accessToken: data['access_token'],
        expiresIn: data['expires_in'],
        headers: sessionHeaders,
      );
    } else {
      throw Exception('Failed to get anonymous token: ${response.statusCode} ${response.body}');
    }
  }

  /// This method gets a new anonymous token. It does not "refresh" the existing
  /// token in the traditional OAuth2 sense. This behavior is consistent with
  /// the original redlib implementation, which also gets a new token each time.
  static Future<OauthSession> refreshToken(OauthSession oldSession) async {
    return login();
  }
}
