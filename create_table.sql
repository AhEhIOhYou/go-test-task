CREATE TABLE comments_statistics (
	post_id INT,
	word TEXT,
	count INT,
	UNIQUE(post_id, word)
);
