import sys
import json
from letterboxdpy.search import Search

def search_for_lists(query):
    """
    Searches for Letterboxd lists based on a query.
    """
    try:
        search_instance = Search(query, "lists")
        search_data = search_instance.results
    except Exception as e:
        return {"error": f"Failed to search for lists: {e}"}

    if not search_data.get('available'):
        return []

    results_out = []
    for result in search_data.get('results', []):
        try:
            results_out.append({
                "name": result.get('title', 'Untitled List'), 
                "owner": result.get('owner', {}).get('username', 'N/A'),
                "slug": result.get('slug', '')
            })
        except Exception:
            continue
            
    return results_out

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No search query provided"}))
        sys.exit(1)

    query = sys.argv[1]
    lists = search_for_lists(query)
    print(json.dumps(lists, indent=4))

