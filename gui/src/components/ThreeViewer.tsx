import { useEffect, useRef } from 'react';
import * as THREE from 'three';

interface ThreeViewerProps {
    label: 'old' | 'new';
    color: string;
}

export default function ThreeViewer({ label, color }: ThreeViewerProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const rendererRef = useRef<THREE.WebGLRenderer | null>(null);

    useEffect(() => {
        if (!containerRef.current) return;

        const container = containerRef.current;
        const width = container.clientWidth;
        const height = container.clientHeight;

        // Scene
        const scene = new THREE.Scene();
        scene.background = new THREE.Color(0x0d1117);

        // Camera
        const camera = new THREE.PerspectiveCamera(50, width / height, 0.1, 1000);
        camera.position.set(3, 2.5, 3);
        camera.lookAt(0, 0.5, 0);

        // Renderer
        const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
        renderer.setSize(width, height);
        renderer.setPixelRatio(window.devicePixelRatio);
        container.appendChild(renderer.domElement);
        rendererRef.current = renderer;

        // Lights
        const ambientLight = new THREE.AmbientLight(0x404040, 2);
        scene.add(ambientLight);

        const dirLight = new THREE.DirectionalLight(0xffffff, 1.5);
        dirLight.position.set(5, 5, 5);
        scene.add(dirLight);

        const backLight = new THREE.DirectionalLight(0x4488ff, 0.5);
        backLight.position.set(-5, 2, -5);
        scene.add(backLight);

        // Grid
        const gridHelper = new THREE.GridHelper(6, 12, 0x30363d, 0x21262d);
        scene.add(gridHelper);

        // Create a demo CAD-like object
        const mainColor = new THREE.Color(color);

        // Main body (box)
        const bodyGeometry = new THREE.BoxGeometry(1.5, 0.8, 1);
        const bodyMaterial = new THREE.MeshPhongMaterial({
            color: mainColor,
            transparent: true,
            opacity: 0.85,
            shininess: 80,
        });
        const body = new THREE.Mesh(bodyGeometry, bodyMaterial);
        body.position.y = 0.4;
        scene.add(body);

        // Edges
        const edges = new THREE.EdgesGeometry(bodyGeometry);
        const edgeMaterial = new THREE.LineBasicMaterial({ color: mainColor, linewidth: 1, transparent: true, opacity: 0.6 });
        const edgeLines = new THREE.LineSegments(edges, edgeMaterial);
        edgeLines.position.copy(body.position);
        scene.add(edgeLines);

        if (label === 'new') {
            // Add a cylinder to "new" version to show the diff
            const cylGeometry = new THREE.CylinderGeometry(0.25, 0.25, 0.6, 16);
            const cylMaterial = new THREE.MeshPhongMaterial({
                color: 0x3fb950,
                transparent: true,
                opacity: 0.85,
                shininess: 80,
                emissive: 0x1a4d2e,
                emissiveIntensity: 0.3,
            });
            const cylinder = new THREE.Mesh(cylGeometry, cylMaterial);
            cylinder.position.set(0.4, 1.1, 0);
            scene.add(cylinder);

            const cylEdges = new THREE.EdgesGeometry(cylGeometry);
            const cylEdgeMat = new THREE.LineBasicMaterial({ color: 0x3fb950, transparent: true, opacity: 0.6 });
            const cylEdgeLines = new THREE.LineSegments(cylEdges, cylEdgeMat);
            cylEdgeLines.position.copy(cylinder.position);
            scene.add(cylEdgeLines);

            // Sphere
            const sphereGeo = new THREE.SphereGeometry(0.2, 16, 16);
            const sphereMat = new THREE.MeshPhongMaterial({
                color: 0x3fb950,
                transparent: true,
                opacity: 0.85,
                emissive: 0x1a4d2e,
                emissiveIntensity: 0.3,
            });
            const sphere = new THREE.Mesh(sphereGeo, sphereMat);
            sphere.position.set(-0.5, 1.0, 0.3);
            scene.add(sphere);
        } else {
            // "old" version: show a piece that will be removed (red highlight)
            const removeGeo = new THREE.BoxGeometry(0.4, 0.3, 0.4);
            const removeMat = new THREE.MeshPhongMaterial({
                color: 0xf85149,
                transparent: true,
                opacity: 0.7,
                emissive: 0x5e1a17,
                emissiveIntensity: 0.4,
            });
            const removeBlock = new THREE.Mesh(removeGeo, removeMat);
            removeBlock.position.set(0.6, 0.95, 0);
            scene.add(removeBlock);

            const remEdges = new THREE.EdgesGeometry(removeGeo);
            const remEdgeMat = new THREE.LineBasicMaterial({ color: 0xf85149, transparent: true, opacity: 0.8 });
            const remEdgeLines = new THREE.LineSegments(remEdges, remEdgeMat);
            remEdgeLines.position.copy(removeBlock.position);
            scene.add(remEdgeLines);
        }

        // Animation
        let angle = 0;
        const animate = () => {
            requestAnimationFrame(animate);
            angle += 0.005;
            camera.position.x = 3 * Math.cos(angle);
            camera.position.z = 3 * Math.sin(angle);
            camera.lookAt(0, 0.5, 0);
            renderer.render(scene, camera);
        };
        animate();

        // Resize
        const handleResize = () => {
            const w = container.clientWidth;
            const h = container.clientHeight;
            camera.aspect = w / h;
            camera.updateProjectionMatrix();
            renderer.setSize(w, h);
        };
        window.addEventListener('resize', handleResize);

        return () => {
            window.removeEventListener('resize', handleResize);
            renderer.dispose();
            if (container.contains(renderer.domElement)) {
                container.removeChild(renderer.domElement);
            }
        };
    }, [label, color]);

    return (
        <div className="viewer-3d-panel">
            <div className={`viewer-3d-label ${label}`}>
                {label === 'old' ? '◀ Before' : 'After ▶'}
            </div>
            <div ref={containerRef} className="viewer-3d-canvas" />
        </div>
    );
}
