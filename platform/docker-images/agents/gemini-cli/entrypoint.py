import os
import sys
import time
from google import genai

def main():
    print("Gemini CLI Agent Started")
    print("------------------------")
    
    api_key = os.environ.get("GEMINI_API_KEY")
    if not api_key:
        print("Error: GEMINI_API_KEY environment variable not set")
        # Keep running to allow debugging/exec
        while True:
            time.sleep(60)
            
    client = genai.Client(api_key=api_key)
    
    print(f"Initialized Gemini client. Ready to process requests.")
    
    # Simple REPL loop for now
    while True:
        try:
            user_input = input("gemini> ")
            if user_input.lower() in ["exit", "quit"]:
                break
                
            if not user_input.strip():
                continue
                
            response = client.models.generate_content(
                model="gemini-2.0-flash",
                contents=user_input
            )
            
            print(response.text)
            
        except KeyboardInterrupt:
            print("\nExiting...")
            break
        except Exception as e:
            print(f"Error: {e}")

if __name__ == "__main__":
    main()

