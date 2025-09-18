class Comment {
  final String id;
  final String author;
  final String body;
  final int score;
  final List<Comment> replies;

  Comment({
    required this.id,
    required this.author,
    required this.body,
    required this.score,
    required this.replies,
  });

  factory Comment.fromJson(Map<String, dynamic> json) {
    final data = json['data'];
    if (data == null) {
      return Comment(id: '', author: '', body: '', score: 0, replies: []);
    }

    final repliesData = data['replies'];
    List<Comment> replies = [];
    if (repliesData != null && repliesData is Map && repliesData.containsKey('data')) {
        final children = repliesData['data']['children'];
        if (children is List) {
            replies = children.where((reply) => reply['kind'] == 't1').map((reply) => Comment.fromJson(reply)).toList();
        }
    }

    return Comment(
      id: data['id'] ?? '',
      author: data['author'] ?? '[deleted]',
      body: data['body'] ?? '',
      score: data['score'] ?? 0,
      replies: replies,
    );
  }
}
