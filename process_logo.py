import sys
from PIL import Image

def process_image(input_path, output_path):
    img = Image.open(input_path).convert("RGBA")
    data = img.getdata()
    
    new_data = []
    # Replace white (#FFFFFF) background with transparent
    for item in data:
        # Check if pixel is white or very close to white
        if item[0] > 240 and item[1] > 240 and item[2] > 240:
            new_data.append((255, 255, 255, 0))
        else:
            new_data.append(item)
            
    img.putdata(new_data)
    
    # getbbox() finds the bounding box of non-zero alpha pixels
    bbox = img.getbbox()
    if bbox:
        img = img.crop(bbox)
        
    width, height = img.size
    size = max(width, height)
    
    # Pad to exact square with transparent background
    final = Image.new("RGBA", (size, size), (255, 255, 255, 0))
    final.paste(img, ((size - width) // 2, (size - height) // 2))
    final.save(output_path)
    print("Cleaned logo saved to", output_path)

if __name__ == '__main__':
    process_image('/Users/gurbajsingh/Documents/gitcad/logo.png', '/tmp/logo_clean.png')
