import 'dart:async';
import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'oauth.dart';

class SessionManager {
  static final SessionManager _instance = SessionManager._internal();
  factory SessionManager() => _instance;
  SessionManager._internal();

  OauthSession? _session;
  DateTime? _sessionExpiry;

  static const _sessionKey = 'oauth_session';

  Future<OauthSession> getSession() async {
    if (_session != null && _sessionExpiry != null && DateTime.now().isBefore(_sessionExpiry!)) {
      return _session!;
    }

    final prefs = await SharedPreferences.getInstance();
    final sessionJson = prefs.getString(_sessionKey);

    if (sessionJson != null) {
      try {
        final sessionData = json.decode(sessionJson);
        final expiry = DateTime.parse(sessionData['expiry']);
        if (DateTime.now().isBefore(expiry)) {
          _session = OauthSession(
            accessToken: sessionData['access_token'],
            expiresIn: sessionData['expires_in'],
            headers: Map<String, String>.from(sessionData['headers']),
          );
          _sessionExpiry = expiry;
          return _session!;
        }
      } catch (e) {
        // Corrupted session data, get a new one
      }
    }

    // No valid session, get a new one
    final newSession = await Oauth.login();
    await _saveSession(newSession);
    return newSession;
  }

  Future<void> _saveSession(OauthSession session) async {
    _session = session;
    // Refresh token a bit before it expires
    _sessionExpiry = DateTime.now().add(Duration(seconds: session.expiresIn - 120));

    final prefs = await SharedPreferences.getInstance();
    final sessionData = {
      'access_token': session.accessToken,
      'expires_in': session.expiresIn,
      'headers': session.headers,
      'expiry': _sessionExpiry!.toIso8601String(),
    };
    await prefs.setString(_sessionKey, json.encode(sessionData));
  }

  Future<void> clearSession() async {
    _session = null;
    _sessionExpiry = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_sessionKey);
  }
}
