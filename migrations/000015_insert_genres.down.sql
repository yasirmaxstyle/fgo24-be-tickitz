DELETE FROM genres
WHERE
    name IN (
        'Action',
        'Comedy',
        'Crime',
        'Drama',
        'Horror',
        'Romance',
        'Sci-Fi',
        'Thriller',
        'Adventure',
        'Animation',
        'Documentary',
        'Fantasy',
        'Mystery',
        'Family',
        'History',
        'Music',
        'Thriller',
        'War'
    );