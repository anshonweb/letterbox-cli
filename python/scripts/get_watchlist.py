#!/usr/bin/env python3


import sys


import json


from letterboxdpy.user import User





def get_watchlist(username):


    """Fetches the watchlist for a given Letterboxd username."""


    try:


        user_instance = User(username)


        watchlist_data = user_instance.get_watchlist()





        if not watchlist_data or not isinstance(watchlist_data, dict) or not watchlist_data.get('available'):


            return []


        movie_dict = watchlist_data.get('data', {})


        if not movie_dict or not isinstance(movie_dict, dict):


            return []





        watchlist_movies = []


        for movie_info in movie_dict.values():





            watchlist_movies.append({


                "title": movie_info.get('name', 'Untitled'),


                "year": movie_info.get('year', 0),


                "slug": movie_info.get('slug', '')


            })





        return watchlist_movies





    except Exception as e:


        return {"error": f"Failed to fetch watchlist for '{username}'. Exception: {e}"}





if __name__ == "__main__":


    if len(sys.argv) < 2:


        print(json.dumps({"error": "No username provided"}))


        sys.exit(1)





    username = sys.argv[1]


    watchlist = get_watchlist(username)


    if isinstance(watchlist, dict) and "error" in watchlist:


         print(json.dumps(watchlist))


         sys.exit(1) 





    print(json.dumps(watchlist, indent=4))