import json
import sys
from letterboxdpy.user import User

def user_details(username):
    try:
        user_instance = User(username)

        movie_name = "N/A"
        recent_movies = []
        try:
            diary = user_instance.get_diary_recent()
            if diary and diary.get('months'):
                months_dict = diary['months']
                for month_key in sorted(months_dict.keys(), reverse=True):
                    days_dict = months_dict[month_key]
                    for day_key in sorted(days_dict.keys(), key=int, reverse=True):
                        for movie in days_dict[day_key]:
                            if 'name' in movie:
                                recent_movies.append(movie['name'])
                if recent_movies:
                    movie_name = recent_movies[0]
        except Exception:
            pass

        followers_list = []
        try:
            followers_list = [user_info.get('name', 'Unknown') for user_info in user_instance.get_followers().values()]
        except Exception:
            pass

        following_list = []
        try:
            following_list = [user_info.get('name', 'Unknown') for user_info in user_instance.get_following().values()]
        except Exception:
            pass
            

        films_watched_count = 0
        try:
            films_data = user_instance.get_films()
            if films_data and 'count' in films_data:
                films_watched_count = films_data['count']
        except Exception:
            pass

        reviews_list = []
        try:
            review_data = user_instance.get_reviews()
            if review_data and 'reviews' in review_data:
                for review in review_data['reviews'].values():
                    if 'movie' in review and 'review' in review and 'date' in review:
                        reviews_list.append({
                            "movie_name": review['movie'].get('name', 'N/A'),
                            "movie_year": review['movie'].get('release', 0),
                            "rating": review.get('rating', 0),
                            "review_text": review['review'].get('content', ''),
                            "review_date": f"{review['date'].get('year', 0)}-{review['date'].get('month', 0):02d}-{review['date'].get('day', 0):02d}"
                        })
        except Exception:
            pass

        return {
            "username": user_instance.username,
            "films_watched": films_watched_count,
            "bio": user_instance.bio,
            "following": following_list,
            "followers": followers_list,
            "favorites": [movie_info.get('name', 'Untitled') for movie_info in user_instance.favorites.values()],
            "last_watched": movie_name,
            "reviews": reviews_list,
            "recent": recent_movies,
            "this_year": user_instance.get_stats()['this_year'],
            "website": user_instance.website,
            "location": user_instance.location,
        }
    
    except Exception as e:
        return {"error": f"Failed to fetch user '{username}'. Exception: {e}"}

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No username provided"}))
        sys.exit(1)

    username = sys.argv[1]
    details = user_details(username)
    print(json.dumps(details, indent=4))

