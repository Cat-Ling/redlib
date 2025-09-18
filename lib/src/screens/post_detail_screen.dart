import 'package:flutter/material.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
import 'package:provider/provider.dart';
import '../client/reddit_client.dart';
import '../models/comment.dart';
import '../models/post.dart';

class PostDetailScreen extends StatefulWidget {
  final Post post;

  const PostDetailScreen({Key? key, required this.post}) : super(key: key);

  @override
  _PostDetailScreenState createState() => _PostDetailScreenState();
}

class _PostDetailScreenState extends State<PostDetailScreen> {
  late Future<List<Comment>> _commentsFuture;

  @override
  void initState() {
    super.initState();
    _commentsFuture = _fetchComments();
  }

  Future<List<Comment>> _fetchComments() async {
    final client = Provider.of<RedditClient>(context, listen: false);
    // The first element of the response is the post listing, the second is the comment listing.
    final data = await client.getComments(widget.post.subreddit, widget.post.id);
    final commentsData = data[1]['data']['children'] as List;
    return commentsData
        .where((item) => item['kind'] == 't1') // Filter out non-comment items
        .map((item) => Comment.fromJson(item))
        .toList();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text("r/${widget.post.subreddit}"),
      ),
      body: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(12.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                widget.post.title,
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              SizedBox(height: 8),
              Text('u/${widget.post.author} • ${widget.post.score} upvotes'),
              SizedBox(height: 12),
              if (widget.post.selftext.isNotEmpty)
                MarkdownBody(data: widget.post.selftext),
              Divider(),
              Text(
                'Comments',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              SizedBox(height: 8),
              _buildCommentsSection(),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCommentsSection() {
    return FutureBuilder<List<Comment>>(
      future: _commentsFuture,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return Center(child: CircularProgressIndicator());
        } else if (snapshot.hasError) {
          return Center(child: Text('Error: ${snapshot.error}'));
        } else if (!snapshot.hasData || snapshot.data!.isEmpty) {
          return Center(child: Text('No comments found.'));
        }

        final comments = snapshot.data!;
        return ListView.builder(
          shrinkWrap: true,
          physics: NeverScrollableScrollPhysics(),
          itemCount: comments.length,
          itemBuilder: (context, index) {
            return CommentWidget(comment: comments[index]);
          },
        );
      },
    );
  }
}

class CommentWidget extends StatelessWidget {
  final Comment comment;
  final int depth;

  const CommentWidget({Key? key, required this.comment, this.depth = 0})
      : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(left: depth * 12.0, top: 8.0, bottom: 8.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Text(
                'u/${comment.author}',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              SizedBox(width: 8),
              Text('${comment.score} points'),
            ],
          ),
          SizedBox(height: 4),
          MarkdownBody(data: comment.body),
          if (comment.replies.isNotEmpty)
            ListView.builder(
              shrinkWrap: true,
              physics: NeverScrollableScrollPhysics(),
              itemCount: comment.replies.length,
              itemBuilder: (context, index) {
                return CommentWidget(
                  comment: comment.replies[index],
                  depth: depth + 1,
                );
              },
            ),
        ],
      ),
    );
  }
}
