#!/usr/bin/env python3
import sys
import json
import re
import requests
from bs4 import BeautifulSoup
from urllib.parse import urlparse, parse_qs
from letterboxdpy.movie import Movie


def convert_stars_to_float(star_string):
    if not star_string:
        return 0.0
    rating = float(star_string.count('★'))
    if '½' in star_string:
        rating += 0.5
    return rating


def get_watch_providers(slug):
    lb_url = f"https://letterboxd.com/film/{slug}/"
    headers = {"User-Agent": "Mozilla/5.0"}

    try:
        res = requests.get(lb_url, headers=headers, timeout=10)
        res.raise_for_status()
    except requests.RequestException as e:
        return []

    match = re.search(r'https://www\.themoviedb\.org/movie/(\d+)', res.text)
    if not match:
        return []

    tmdb_id = match.group(1)
    tmdb_url = f"https://www.themoviedb.org/movie/{tmdb_id}/watch"

    try:
        res = requests.get(tmdb_url, headers=headers, timeout=10)
        res.raise_for_status()
    except requests.RequestException:
        return []

    soup = BeautifulSoup(res.text, "html.parser")
    unique_providers = {}

    name_corrections = {
        "JioHotstar": "Disney+ Hotstar"
    }

    for a in soup.select(".ott_provider a"):
        title = a.get("title")
        link = a.get("href")
        if not title or not link:
            continue

        action_type = "unknown"
        if title.lower().startswith("watch "):
            action_type = "stream"
        elif title.lower().startswith("buy "):
            action_type = "buy"
        elif title.lower().startswith("rent "):
            action_type = "rent"
        name_match = re.search(r' on (.+)$', title)
        if not name_match:
            continue
        provider_name = name_match.group(1).strip()
        provider_name = name_corrections.get(provider_name, provider_name)
        final_link = link
        if "click.justwatch.com" in link:
            parsed_url = urlparse(link)
            query_params = parse_qs(parsed_url.query)
            if 'r' in query_params:
                final_link = query_params['r'][0]
        if provider_name not in unique_providers or action_type == "stream":
            unique_providers[provider_name] = {
                "type": action_type,
                "link": final_link
            }
    providers = [
        {"name": name, "type": data["type"], "link": data["link"]}
        for name, data in unique_providers.items()
    ]

    return providers


def format_runtime(total_minutes):
    if not total_minutes or not isinstance(total_minutes, int) or total_minutes <= 0:
        return None
    hours, minutes = divmod(total_minutes, 60)
    return f"{hours}h {minutes}min"


def get_movie_details(slug):
    try:
        movie_instance = Movie(slug)
        similar_list = []
        similar_data = movie_instance.get_similar_movies()
        if similar_data:
            for data in similar_data.values():
                similar_list.append({
                    "name": data.get("name"),
                    "rating": data.get("rating", 0.0)
                })

        return {
            "title": movie_instance.title,
            "year": movie_instance.year,
            "director": movie_instance.crew.get("director", [{}])[0].get("name", "Unknown"),
            "genres": [item['name'] for item in movie_instance.genres if item.get('type') == 'genre'],
            "rating": movie_instance.rating,
            "description": movie_instance.description,
            "url": movie_instance.url,
            "reviews": [
                {
                    "author": review['user']['username'],
                    "text": review.get('review', '').strip(),
                    "rating": convert_stars_to_float(review.get('rating'))
                }
                for review in getattr(movie_instance, "popular_reviews", [])[:5]
            ],
            "providers": get_watch_providers(slug),
            "runtime": format_runtime(movie_instance.runtime),
            "cast" : [actor['name'] for actor in movie_instance.cast[:5]],
        
            "release_date": movie_instance.details,
            
            "members": movie_instance.get_watchers_stats()['members'],
            "fans":  movie_instance.get_watchers_stats()['fans'],
            "likes":  movie_instance.get_watchers_stats()['likes'],
            "review_count":  movie_instance.get_watchers_stats()['reviews'],
            "lists":  movie_instance.get_watchers_stats()['lists'],
            "tagline": movie_instance.get_tagline(),
            "similar": similar_list
        }
    except Exception as e:
        return {"error": str(e)}


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No slug provided"}))
        sys.exit(1)

    slug = sys.argv[1]
    details = get_movie_details(slug)
    print(json.dumps(details, indent=4))

