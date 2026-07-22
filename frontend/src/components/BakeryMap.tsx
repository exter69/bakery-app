import { useEffect, useState } from 'react';
import { MapContainer, TileLayer, Marker, Popup, Circle, useMap } from 'react-leaflet';
import L from 'leaflet';
import { Link } from 'react-router-dom';
import 'leaflet/dist/leaflet.css';
import type { BakeryCard } from '../types/bakery';

// Fix default marker icons for bundlers
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png';
import markerIcon from 'leaflet/dist/images/marker-icon.png';
import markerShadow from 'leaflet/dist/images/marker-shadow.png';

delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
});

const bakeryIcon = new L.Icon({
  iconUrl: markerIcon,
  iconRetinaUrl: markerIcon2x,
  shadowUrl: markerShadow,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
});

const RANGE_KM = 20;
const RANGE_METERS = RANGE_KM * 1000;

interface BakeryMapProps {
  bakeries: BakeryCard[];
}

function RecenterMap({ lat, lng }: { lat: number; lng: number }) {
  const map = useMap();
  useEffect(() => {
    map.setView([lat, lng], 12);
  }, [map, lat, lng]);
  return null;
}

export default function BakeryMap({ bakeries }: BakeryMapProps) {
  const [userPos, setUserPos] = useState<{ lat: number; lng: number } | null>(null);

  useEffect(() => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (pos) => setUserPos({ lat: pos.coords.latitude, lng: pos.coords.longitude }),
        () => setUserPos({ lat: 50.8503, lng: 4.3517 }) // Default Brussels
      );
    } else {
      setUserPos({ lat: 50.8503, lng: 4.3517 });
    }
  }, []);

  if (!userPos) {
    return (
      <div style={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#7a5c3e', fontFamily: 'var(--font-hand-body)' }}>
        Loading map…
      </div>
    );
  }

  // Filter bakeries within 20km range
  const nearby = bakeries.filter((b) => {
    if (!b.latitude && !b.longitude) return false;
    if (b.distance != null) return b.distance <= RANGE_KM;
    return true;
  });

  return (
    <div style={{ borderRadius: '12px', overflow: 'hidden', border: '1.5px solid #3a2e22', boxShadow: '3px 3px 0 rgba(58,46,34,0.12)' }}>
      <MapContainer
        center={[userPos.lat, userPos.lng]}
        zoom={12}
        style={{ height: 320, width: '100%' }}
        scrollWheelZoom={false}
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <RecenterMap lat={userPos.lat} lng={userPos.lng} />

        {/* 20km range circle */}
        <Circle
          center={[userPos.lat, userPos.lng]}
          radius={RANGE_METERS}
          pathOptions={{ color: '#e8b04b', fillColor: '#e8b04b', fillOpacity: 0.06, weight: 1.5, dashArray: '6 4' }}
        />

        {/* Bakery markers */}
        {nearby.map((bakery) => (
          <Marker key={bakery.id} position={[bakery.latitude, bakery.longitude]} icon={bakeryIcon}>
            <Popup>
              <div style={{ fontFamily: 'sans-serif', fontSize: '13px' }}>
                <strong>{bakery.name}</strong>
                <br />
                {bakery.todaySchedule.isOpen
                  ? `Open ${bakery.todaySchedule.openTime} – ${bakery.todaySchedule.closeTime}`
                  : 'Closed today'}
                {bakery.distance != null && (
                  <>
                    <br />
                    {bakery.distance < 1
                      ? `${Math.round(bakery.distance * 1000)}m away`
                      : `${bakery.distance} km away`}
                  </>
                )}
                <br />
                <Link to={`/bakeries/${bakery.id}`} style={{ color: '#e8b04b', fontWeight: 600 }}>
                  View →
                </Link>
              </div>
            </Popup>
          </Marker>
        ))}
      </MapContainer>
    </div>
  );
}
