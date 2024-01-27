import json
from ultralytics import YOLO
import cv2
import numpy as np
from matplotlib.colors import to_hex

async def exportData(image):
    model = YOLO("yolo.pt")
    img = cv2.imdecode(np.frombuffer(image, np.uint8), cv2.IMREAD_COLOR)
    results = model(img)
    object_info = {}
    json_object = json.loads(results[0].tojson())

    def crop_center(img, box, crop_size):
        x1, y1, x2, y2 = box.values()
        x1 = float(x1)
        y1 = float(y1)
        x2 = float(x2)
        y2 = float(y2)
        center_x = (x1 + x2) / 2
        center_y = (y1 + y2) / 2
        crop_half_size = crop_size // 2
        crop_x1 = int(center_x - crop_half_size)
        crop_x2 = int(center_x + crop_half_size)
        crop_y1 = int(center_y - crop_half_size)
        crop_y2 = int(center_y + crop_half_size)
        return img[crop_y1:crop_y2, crop_x1:crop_x2]

    for i, detection in enumerate(json_object):
        label = detection['name']
        box = detection['box']
        cropped_object = crop_center(img, box, crop_size=20)
        pixels = cropped_object.reshape((-1, 3))
        pixels = np.float32(pixels)

        criteria = (cv2.TERM_CRITERIA_EPS + cv2.TERM_CRITERIA_MAX_ITER, 100, 0.2)
        k = 3 
        _, labels, centers = cv2.kmeans(pixels, k, None, criteria, 10, cv2.KMEANS_RANDOM_CENTERS)
        centers = np.uint8(centers)

        dominant_color_bgr = centers[np.argmax(np.bincount(labels.flatten()))]
        dominant_color_rgb_normalized = (dominant_color_bgr[2] / 255, dominant_color_bgr[1] / 255, dominant_color_bgr[0] / 255)
        dominant_color_hex = to_hex(dominant_color_rgb_normalized)

        object_info[f"Object_{i + 1}"] = {
            'Label': label,
            'Dominant_Color_Hex': dominant_color_hex
        }
    return object_info
