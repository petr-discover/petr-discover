from fastapi import FastAPI

app = FastAPI()


@app.get("/api/v1/ai")
async def root():
    return {"message": "Hello World"}