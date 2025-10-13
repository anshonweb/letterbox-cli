#!/usr/bin/env python3
import sys
import json
from letterboxdpy.movie import Movie


def convert_stars_to_float(star_string):

    if not star_string:
        return 0.0
    rating = float(star_string.count('★'))

    if '½' in star_string:
        rating += 0.5
        
    return rating
def get_movie_details(slug):
    movie_instance = Movie(slug)
    
    
    return {
        "title": movie_instance.title,
        "year": movie_instance.year,
        "director": movie_instance.crew["director"][0]['name'],
        "genres": [item['name'] for item in movie_instance.genres if item['type'] == 'genre'],
        "rating": movie_instance.rating,
        "description": movie_instance.description,
        "url": movie_instance.url,
        "reviews": [
    {
        "author": review['user']['username'],
        "text": review['review'],
        "rating": convert_stars_to_float(review.get('rating'))
    }
    for review in movie_instance.popular_reviews
]
    }

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({}))
        sys.exit(0)

    slug = sys.argv[1]
    details = get_movie_details(slug)
    print(json.dumps(details))
