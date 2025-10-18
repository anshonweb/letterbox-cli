import json
import sys
from letterboxdpy.user import User

def user_details(username):
    try:
        user_instance = User(username)

        movie_name = "N/A"
        try:
            diary = user_instance.get_diary_recent()
            if diary and diary.get('months'):
                months_dict = diary['months']
                if months_dict:
                    recent_month_num = max(months_dict.keys())
                    days_dict = months_dict[recent_month_num]
                    if days_dict:
                        recent_day_key = max(days_dict.keys(), key=int)
                        recent_movies_list = days_dict[recent_day_key]
                        if recent_movies_list:
                            recent_movie = recent_movies_list[0]
                            movie_name = recent_movie.get('name', "N/A")
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
            user_reviews_data = user_instance.get_reviews()
            if user_reviews_data and 'reviews' in user_reviews_data:
                for review_data in user_reviews_data['reviews'].values():
                    review_date_obj = review_data.get('date', {})
                    review_date_str = "0000-00-00"
                    if review_date_obj:
                        year = review_date_obj.get('year', 0)
                        month = review_date_obj.get('month', 0)
                        day = review_date_obj.get('day', 0)
                        review_date_str = f"{year:04d}-{month:02d}-{day:02d}"

                    reviews_list.append({
                        "movie_name": review_data.get('movie', {}).get('name', 'Untitled'),
                        "movie_year": review_data.get('movie', {}).get('release', 0),
                        "rating": review_data.get('rating', 0),
                        "review_text": review_data.get('review', {}).get('content', ''),
                        "review_date": review_date_str,
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

