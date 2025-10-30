#!/usr/bin/env python3
import sys
import json
from letterboxdpy.user import User
from datetime import datetime

def get_diary_entries(username):
    """Fetches and formats diary entries for a given Letterboxd username."""
    entries = []
    try:
        user_instance = User(username)
        diary_data = user_instance.get_diary()

        if not diary_data or not isinstance(diary_data, dict) or 'entries' not in diary_data:
            return []

        diary_entries_dict = diary_data.get('entries', {})
        if not diary_entries_dict or not isinstance(diary_entries_dict, dict):
            return []

        for entry_details in diary_entries_dict.values():
            if not isinstance(entry_details, dict): continue

            date_info = entry_details.get('date', {})
            watch_date_str = "Unknown Date"
            if isinstance(date_info, dict):
                try:
                    year = date_info.get('year')
                    month = date_info.get('month')
                    day = date_info.get('day')
                    if year and month and day:
                        watch_date_str = f"{int(year)}-{int(month):02d}-{int(day):02d}"
                        # Validate date
                        datetime.strptime(watch_date_str, '%Y-%m-%d')
                except (ValueError, TypeError):
                    watch_date_str = "Unknown Date"
            actions_info = entry_details.get('actions', {})
            rating_val = 0.0
            rewatch_val = False
            if isinstance(actions_info, dict):
                rating_int = actions_info.get('rating')
                if isinstance(rating_int, (int, float)):
                   rating_val = float(rating_int) / 2.0
                rewatch_val = actions_info.get('rewatched', False)


            entries.append({
                "title": entry_details.get('name', 'Untitled'),
                "year": entry_details.get('release', 0), 
                "rating": rating_val, 
                "watch_date": watch_date_str,
                "rewatch": rewatch_val, 
                "slug": entry_details.get('slug', '')
            })
        entries.sort(key=lambda x: x['watch_date'] if x['watch_date'] != "Unknown Date" else "0000-00-00", reverse=True)

        return entries

    except Exception as e:
        return {"error": f"Failed to fetch diary for '{username}'. Exception: {e}"}

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No username provided"}))
        sys.exit(1)

    username = sys.argv[1]
    diary = get_diary_entries(username)

    if isinstance(diary, dict) and "error" in diary:
         print(json.dumps(diary))
         sys.exit(1)

    print(json.dumps(diary, indent=4))

