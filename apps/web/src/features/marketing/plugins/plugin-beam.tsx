import { useFrame } from '@react-three/fiber';
import { useMemo, useRef } from 'react';
import * as THREE from 'three';

function createParticles(count: number) {
  const positions = new Float32Array(count * 3);
  for (let index = 0; index < count; index += 1) {
    const radius = Math.random() * 3;
    const angle = Math.random() * Math.PI * 2;
    positions[index * 3] = radius * Math.cos(angle);
    positions[index * 3 + 1] = (Math.random() - 0.5) * 10;
    positions[index * 3 + 2] = radius * Math.sin(angle);
  }
  return positions;
}

export function PluginBeam() {
  const beam = useRef<THREE.Mesh>(null);
  const dust = useRef<THREE.Points>(null);
  const count = 170;
  const positions = useMemo(() => createParticles(count), []);
  const uniforms = useMemo(
    () => ({
      uTime: { value: 0 },
      uColor: { value: new THREE.Color('#ffffff') },
    }),
    [],
  );

  useFrame(({ clock }) => {
    if (beam.current) {
      (beam.current.material as THREE.ShaderMaterial).uniforms.uTime.value =
        clock.elapsedTime;
    }
    if (dust.current) {
      dust.current.rotation.y = clock.elapsedTime * 0.03;
      dust.current.position.y = 4 + Math.sin(clock.elapsedTime * 0.15) * 0.2;
    }
  });

  return (
    <group>
      <mesh ref={beam} position={[0, 8, 0]}>
        <cylinderGeometry args={[0.8, 8, 18, 64, 1, true]} />
        <shaderMaterial
          side={THREE.DoubleSide}
          transparent
          depthWrite={false}
          blending={THREE.AdditiveBlending}
          uniforms={uniforms}
          vertexShader="varying vec2 vUv; void main() { vUv = uv; gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0); }"
          fragmentShader="uniform float uTime; uniform vec3 uColor; varying vec2 vUv; float random(vec2 st) { return fract(sin(dot(st.xy, vec2(12.9898, 78.233))) * 43758.5453); } void main() { float beamLongitudinal = smoothstep(0.0, 0.5, vUv.y); float sourceSoftness = smoothstep(1.0, 0.85, vUv.y); float noise = random(vec2(vUv.x * 20.0, uTime * 0.03)); float ray = smoothstep(0.4, 0.6, noise) * 0.06; float alpha = beamLongitudinal * sourceSoftness * 0.12; alpha += ray * beamLongitudinal * sourceSoftness; float core = smoothstep(0.5, 0.88, vUv.y) * smoothstep(1.0, 0.88, vUv.y) * 0.25; alpha += core; gl_FragColor = vec4(uColor, alpha); }"
        />
      </mesh>
      <points ref={dust} position={[0, 4, 0]}>
        <bufferGeometry>
          <bufferAttribute
            attach="attributes-position"
            count={count}
            args={[positions, 3]}
          />
        </bufferGeometry>
        <pointsMaterial
          size={0.06}
          color="#ffffff"
          transparent
          opacity={0.7}
          sizeAttenuation
          blending={THREE.AdditiveBlending}
        />
      </points>
    </group>
  );
}
