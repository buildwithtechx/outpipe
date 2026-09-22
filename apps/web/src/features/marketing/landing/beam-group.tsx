import { useFrame } from '@react-three/fiber';
import { useMemo, useRef } from 'react';
import * as THREE from 'three';

function createDust(count: number) {
  const positions = new Float32Array(count * 3);
  for (let index = 0; index < count; index += 1) {
    const radius = Math.random() * 4;
    const angle = Math.random() * Math.PI * 2;
    positions[index * 3] = (Math.random() - 0.5) * 16;
    positions[index * 3 + 1] = radius * Math.cos(angle);
    positions[index * 3 + 2] = radius * Math.sin(angle);
  }
  return positions;
}

function Dust() {
  const points = useRef<THREE.Points>(null);
  const count = 300;
  const positions = useMemo(() => createDust(count), []);

  useFrame(({ clock }) => {
    if (!points.current) return;
    points.current.rotation.x = clock.elapsedTime * 0.05;
    points.current.position.x = 8 + Math.sin(clock.elapsedTime * 0.2) * 0.5;
  });

  return (
    <points ref={points} position={[8, 0, 0]}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          count={count}
          args={[positions, 3]}
        />
      </bufferGeometry>
      <pointsMaterial
        size={0.05}
        color="#ffffff"
        transparent
        opacity={0.6}
        sizeAttenuation
        blending={THREE.AdditiveBlending}
      />
    </points>
  );
}

function RelayBeam({ color }: { color: string }) {
  const mesh = useRef<THREE.Mesh>(null);
  const uniforms = useMemo(
    () => ({
      uTime: { value: 0 },
      uColor: { value: new THREE.Color(color) },
    }),
    [color],
  );

  useFrame(({ clock }) => {
    if (mesh.current) {
      (mesh.current.material as THREE.ShaderMaterial).uniforms.uTime.value =
        clock.elapsedTime;
    }
  });

  return (
    <mesh ref={mesh} position={[8, 0, 0]} rotation={[0, 0, Math.PI / 2]}>
      <cylinderGeometry args={[1.2, 5, 16, 64, 1, true]} />
      <shaderMaterial
        side={THREE.DoubleSide}
        transparent
        depthWrite={false}
        blending={THREE.AdditiveBlending}
        uniforms={uniforms}
        vertexShader="varying vec2 vUv; void main() { vUv = uv; gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0); }"
        fragmentShader="uniform float uTime; uniform vec3 uColor; varying vec2 vUv; float random(vec2 st) { return fract(sin(dot(st.xy, vec2(12.9898, 78.233))) * 43758.5453); } void main() { float beam = smoothstep(0.0, 0.6, vUv.y); float softness = smoothstep(1.0, 0.85, vUv.y); float ray = smoothstep(0.4, 0.6, random(vec2(vUv.x * 20.0, 0.0))) * 0.04; float alpha = beam * softness * 0.06; alpha += ray * beam * softness; float core = smoothstep(0.5, 0.85, vUv.y) * smoothstep(1.0, 0.85, vUv.y) * 0.15; gl_FragColor = vec4(uColor, alpha + core); }"
      />
    </mesh>
  );
}

export function BeamGroup({ color = '#ffffff' }: { color?: string }) {
  return (
    <group position={[-10, -1.2, 0]} rotation={[0, 0, 0.18]}>
      <RelayBeam color={color} />
      <Dust />
    </group>
  );
}
