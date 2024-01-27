from fastapi import FastAPI, File, UploadFile
from fastapi.responses import FileResponse
from yolo import exportData
import cv2
import numpy as np
from io import BytesIO

app = FastAPI()

@app.post("/api/v1/ai")
async def root(file: UploadFile = File(...)):
    file_contents = await file.read()
    processed_image = await exportData(file_contents)
    return processed_image
