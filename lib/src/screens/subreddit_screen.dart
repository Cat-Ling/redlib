import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../client/reddit_client.dart';
import '../models/post.dart';
import 'post_detail_screen.dart';

class SubredditScreen extends StatefulWidget {
  final String subreddit;

  const SubredditScreen({Key? key, required this.subreddit}) : super(key: key);

  @override
  _SubredditScreenState createState() => _SubredditScreenState();
}

class _SubredditScreenState extends State<SubredditScreen> {
  late Future<List<Post>> _postsFuture;

  @override
  void initState() {
    super.initState();
    _postsFuture = _fetchPosts();
  }

  Future<List<Post>> _fetchPosts() async {
    final client = Provider.of<RedditClient>(context, listen: false);
    final data = await client.getSubreddit(widget.subreddit);
    final posts = (data['data']['children'] as List)
        .map((item) => Post.fromJson(item))
        .toList();
    return posts;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('r/${widget.subreddit}'),
      ),
      body: FutureBuilder<List<Post>>(
        future: _postsFuture,
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return Center(child: CircularProgressIndicator());
          } else if (snapshot.hasError) {
            return Center(child: Text('Error: ${snapshot.error}'));
          } else if (!snapshot.hasData || snapshot.data!.isEmpty) {
            return Center(child: Text('No posts found in r/${widget.subreddit}.'));
          }

          final posts = snapshot.data!;
          return ListView.builder(
            itemCount: posts.length,
            itemBuilder: (context, index) {
              final post = posts[index];
              return ListTile(
                title: Text(post.title),
                subtitle: Text('u/${post.author} • ${post.score} upvotes'),
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
