class Post {
  final String id;
  final String title;
  final String author;
  final String subreddit;
  final int score;
  final String selftext;
  final String url;
  final int numComments;

  Post({
    required this.id,
    required this.title,
    required this.author,
    required this.subreddit,
    required this.score,
    required this.selftext,
    required this.url,
    required this.numComments,
  });

  factory Post.fromJson(Map<String, dynamic> json) {
    final data = json['data'];
    return Post(
      id: data['id'],
      title: data['title'],
      author: data['author'],
      subreddit: data['subreddit'],
      score: data['score'],
      selftext: data['selftext'] ?? '',
      url: data['url'],
      numComments: data['num_comments'],
    );
  }
}
