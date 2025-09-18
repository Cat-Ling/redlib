import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../client/reddit_client.dart';
import '../models/post.dart';
import 'post_detail_screen.dart';
import 'subreddit_screen.dart';

class HomeScreen extends StatefulWidget {
  @override
  _HomeScreenState createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  late Future<List<Post>> _postsFuture;

  @override
  void initState() {
    super.initState();
    // We need to delay the fetch until after the first frame, so that the context is available.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _postsFuture = _fetchPosts();
      setState(() {});
    });
  }

  Future<List<Post>> _fetchPosts() async {
    final client = Provider.of<RedditClient>(context, listen: false);
    final data = await client.getFrontPage();
    final posts = (data['data']['children'] as List)
        .map((item) => Post.fromJson(item))
        .toList();
    return posts;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Reddit'),
      ),
      body: FutureBuilder<List<Post>>(
        future: _postsFuture,
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return Center(child: CircularProgressIndicator());
          } else if (snapshot.hasError) {
            return Center(child: Text('Error: ${snapshot.error}'));
          } else if (!snapshot.hasData || snapshot.data!.isEmpty) {
            return Center(child: Text('No posts found.'));
          }

          final posts = snapshot.data!;
          return ListView.builder(
            itemCount: posts.length,
            itemBuilder: (context, index) {
              final post = posts[index];
              return ListTile(
                title: Text(post.title),
                subtitle: RichText(
                  text: TextSpan(
                    style: Theme.of(context).textTheme.bodySmall,
                    children: [
                      TextSpan(text: 'u/${post.author} • '),
                      TextSpan(
                        text: 'r/${post.subreddit}',
                        style: TextStyle(color: Colors.blue),
                        recognizer: TapGestureRecognizer()
                          ..onTap = () {
                            Navigator.push(
                              context,
                              MaterialPageRoute(
                                builder: (context) => SubredditScreen(subreddit: post.subreddit),
                              ),
                            );
                          },
                      ),
                      TextSpan(text: ' • ${post.score} upvotes'),
                    ],
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => PostDetailScreen(post: post),
                    ),
                  );
                },
              );
            },
          );
        },
      ),
    );
  }
}
