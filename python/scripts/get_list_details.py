import sys
import json
from letterboxdpy.list import List

def get_list_movies(owner_username, list_slug):
    """
    Fetches all movies from a specific Letterboxd list.
    """
    try:
        list_instance = List(owner_username, list_slug)
        movies = list_instance.movies
    except Exception as e:
        return {"error": f"Failed to fetch list '{owner_username}/{list_slug}': {e}"}

    movies_out = []
    for movie_data in movies.values():
        try:
            url = movie_data.get('url', '')
            slug = url.split('/film/')[-1].strip('/')
            
            movies_out.append({
                "title": movie_data.get('name', 'Untitled'),
                "year": movie_data.get('year', 0),
                "slug": slug,
                "director": movie_data.get('director', 'N/A') 
            })
        except Exception:
            continue
            
    return movies_out

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print(json.dumps({"error": "Usage: python get_list_details.py <owner_username> <list_slug>"}))
        sys.exit(1)

    owner = sys.argv[1]
    slug = sys.argv[2]
    details = get_list_movies(owner, slug)
    print(json.dumps(details, indent=4))
