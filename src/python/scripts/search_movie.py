#!/usr/bin/env python3
import sys
import json
from letterboxdpy.search import Search

def search_movie(query):
    search_instance = Search(query, 'films')
    search_data = search_instance.get_results(max=5) 
    
    movies = []
    for item in search_data.get('results', []): 
        if 'name' in item and 'year' in item and 'slug' in item:
            movies.append({
                "title": item['name'],
                "year": item['year'],
                "slug": item['slug'],
                "director": item['directors'][0]['name'],
            })

    return movies

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps([]))
        sys.exit(0)

    query = sys.argv[1]
    movies = search_movie(query)
    print(json.dumps(movies, indent=4)) 